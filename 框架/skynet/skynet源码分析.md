[toc]

## 1. 模块加载

skynet底层代码位于skynet/skynet-src下,模块加载相关在skynet-module.c skynet-module.h这两个文件里。这里的模块在linux下指的是so，在windows下指的是dll，在skynet中指的是config中配置的cpath下的文件。

skynet_module.h源码(skynet-src/skynet_module.h):

```c
#ifndef SKYNET_MODULE_H
#define SKYNET_MODULE_H

struct skynet_context;

typedef void * (*skynet_dl_create)(void);
typedef int (*skynet_dl_init)(void * inst, struct skynet_context *, const char * parm);
typedef void (*skynet_dl_release)(void * inst);
typedef void (*skynet_dl_signal)(void * inst, int signal);

struct skynet_module {
	const char * name;
	void * module;
	skynet_dl_create create;
	skynet_dl_init init;
	skynet_dl_release release;
	skynet_dl_signal signal;
};

void skynet_module_insert(struct skynet_module *mod);
struct skynet_module * skynet_module_query(const char * name);
void * skynet_module_instance_create(struct skynet_module *);
int skynet_module_instance_init(struct skynet_module *, void * inst, struct skynet_context *ctx, const char * parm);
void skynet_module_instance_release(struct skynet_module *, void *inst);
void skynet_module_instance_signal(struct skynet_module *, void *inst, int signal);

void skynet_module_init(const char *path);

#endif

```

skynet_module.c源码(skynet-src/skynet_module.c):

```c
#include "skynet.h"

#include "skynet_module.h"
#include "spinlock.h"

#include <assert.h>
#include <string.h>
#include <dlfcn.h>
#include <stdlib.h>
#include <stdint.h>
#include <stdio.h>

#define MAX_MODULE_TYPE 32

//核心功能是启动c写的so模块

struct modules {
	int count;
	struct spinlock lock;
	const char * path;
	struct skynet_module m[MAX_MODULE_TYPE];
};

static struct modules * M = NULL;

//打开一个动态库
static void *
_try_open(struct modules *m, const char * name) {
	const char *l;
	const char * path = m->path;
	size_t path_size = strlen(path);
	size_t name_size = strlen(name);

	int sz = path_size + name_size;
	//search path
	void * dl = NULL;
	char tmp[sz];
	//遍历路径查找so，以';'分隔
	do
	{
		memset(tmp,0,sz);
		while (*path == ';') path++;
		if (*path == '\0') break;
		l = strchr(path, ';');
		if (l == NULL) l = path + strlen(path);
		int len = l - path;
		int i;
		for (i=0;path[i]!='?' && i < len ;i++) {
			tmp[i] = path[i];
		}
		memcpy(tmp+i,name,name_size);
		if (path[i] == '?') {
			strncpy(tmp+i+name_size,path+i+1,len - i - 1);
		} else {
			fprintf(stderr,"Invalid C service path\n");
			exit(1);
		}
		dl = dlopen(tmp, RTLD_NOW | RTLD_GLOBAL);
		path = l;
	}while(dl == NULL);

	if (dl == NULL) {
		fprintf(stderr, "try open %s failed : %s\n",name,dlerror());
	}

	return dl;
}

static struct skynet_module * 
_query(const char * name) {
	int i;
	for (i=0;i<M->count;i++) {
		if (strcmp(M->m[i].name,name)==0) {
			return &M->m[i];
		}
	}
	return NULL;
}

static void *
get_api(struct skynet_module *mod, const char *api_name) {
	size_t name_size = strlen(mod->name);
	size_t api_size = strlen(api_name);
	char tmp[name_size + api_size + 1];
	memcpy(tmp, mod->name, name_size);
	memcpy(tmp+name_size, api_name, api_size+1);
	char *ptr = strrchr(tmp, '.');
	if (ptr == NULL) {
		ptr = tmp;
	} else {
		ptr = ptr + 1;
	}
	return dlsym(mod->module, ptr);
}

static int
open_sym(struct skynet_module *mod) {
	mod->create = get_api(mod, "_create");
	mod->init = get_api(mod, "_init");
	mod->release = get_api(mod, "_release");
	mod->signal = get_api(mod, "_signal");

	return mod->init == NULL;
}


//查询模块
struct skynet_module * 
skynet_module_query(const char * name) {
	struct skynet_module * result = _query(name);
	if (result)
		return result;

	SPIN_LOCK(M)

	result = _query(name); // double check spin_lock(M)阻塞过程中可能会被写入列表，double check可防止重复写入列表

	//如果不在列表中，打开so
	if (result == NULL && M->count < MAX_MODULE_TYPE) {
		int index = M->count;
		void * dl = _try_open(M,name);
		if (dl) {
			M->m[index].name = name;
			M->m[index].module = dl;

			if (open_sym(&M->m[index]) == 0) {
				M->m[index].name = skynet_strdup(name);
				M->count ++;
				result = &M->m[index];
			}
		}
	}

	SPIN_UNLOCK(M)

	return result;
}

//添加模块到模块列表
void 
skynet_module_insert(struct skynet_module *mod) {
	SPIN_LOCK(M)

	struct skynet_module * m = _query(mod->name);
	assert(m == NULL && M->count < MAX_MODULE_TYPE);	//已经存在，则报错并终止程序
	int index = M->count;
	M->m[index] = *mod;
	++M->count;

	SPIN_UNLOCK(M)
}

//create做内存分配
void * 
skynet_module_instance_create(struct skynet_module *m) {
	if (m->create) {
		return m->create();
	} else {
		return (void *)(intptr_t)(~0);
	}
}

//做初始化
int
skynet_module_instance_init(struct skynet_module *m, void * inst, struct skynet_context *ctx, const char * parm) {
	return m->init(inst, ctx, parm);
}

//资源回收
void 
skynet_module_instance_release(struct skynet_module *m, void *inst) {
	if (m->release) {
		m->release(inst);
	}
}

//发信号
void
skynet_module_instance_signal(struct skynet_module *m, void *inst, int signal) {
	if (m->signal) {
		m->signal(inst, signal);
	}
}

//在skynet-main.c中被调用，传进来的path是在配置文件中配置的capth属性，默认加载cpath目录下的so文件。
void 
skynet_module_init(const char *path) {
	struct modules *m = skynet_malloc(sizeof(*m));
	m->count = 0;
	m->path = skynet_strdup(path);

	SPIN_INIT(m)

	M = m;
}

```

skynet要求模块的create/init/release/signal方法的命名是模块名加一个下划线，后面带create/init/release/signal。

> **总结：模块加载的流程是在config文件中配置一个cpath，它包含了你想要加载的so的路径。然后skynet-main.c在启动的时候会把cpath读出来，设进module_path中。在skynet_server.c中的skynet_context_new中会调用skynet_module_query，skynet_module_query首先会在列表中查询so是否已经加载，如果没有就直接加载它。【引用自[skynet源码分析（1）--模块加载](https://www.jianshu.com/p/1aa8b0fc6c11)】**



## 2. monitor

skynet对服务的监控实现在skynet_monitor.c和skynet_monitor.h中，当服务可能陷入死循环的时候就打一条日志。

skynet_monitor.h源码(skynet-src/skynet_monitor.h):

```c
#ifndef SKYNET_MONITOR_H
#define SKYNET_MONITOR_H

#include <stdint.h>

struct skynet_monitor;

struct skynet_monitor * skynet_monitor_new();
void skynet_monitor_delete(struct skynet_monitor *);
void skynet_monitor_trigger(struct skynet_monitor *, uint32_t source, uint32_t destination);
void skynet_monitor_check(struct skynet_monitor *);

#endif

```

Skynet_monitor.c源码(skynet-src/skynet_monitor.c):

```c
#include "skynet.h"

#include "skynet_monitor.h"
#include "skynet_server.h"
#include "skynet.h"
#include "atomic.h"

#include <stdlib.h>
#include <string.h>



struct skynet_monitor {
	int version;
	int check_version;
	uint32_t source;
	uint32_t destination;
};

//结构体初始化
struct skynet_monitor * 
skynet_monitor_new() {
	struct skynet_monitor * ret = skynet_malloc(sizeof(*ret));
	memset(ret, 0, sizeof(*ret));
	return ret;
}

//结构体回收
void 
skynet_monitor_delete(struct skynet_monitor *sm) {
	skynet_free(sm);
}

//触发监控
//每次消息派发都会调用skynet_monitor_trigger,一共两次，第一次参数source和destination都为真实的值，也就是不为0。
//第二次调用是在消息派发完成的时候，source和destination都赋0。
//如果第一次调用trigger以后，消息派发迟迟无法完成，monitor线程第一次检查，会将check_version的值赋为version
//然后monitor线程第二次检查，这个时候version和check_version就会相等，而且这时候destination也不为0，就会进入释放目标服务和打印报警的流程
void 
skynet_monitor_trigger(struct skynet_monitor *sm, uint32_t source, uint32_t destination) {
	sm->source = source;
	sm->destination = destination;
	ATOM_INC(&sm->version);
}

//检查监控 有一个线程专门做这个
void 
skynet_monitor_check(struct skynet_monitor *sm) {
	if (sm->version == sm->check_version) {
		//第二次检查，如果destination不为0，证明派发未完成
		if (sm->destination) {
			//释放目标服务
			skynet_context_endless(sm->destination);
			//打日志，警告可能死循环
			skynet_error(NULL, "A message from [ :%08x ] to [ :%08x ] maybe in an endless loop (version = %d)", sm->source , sm->destination, sm->version);
		}
	} else {
		//第一次检查，将check_version的值赋为version
		sm->check_version = sm->version;
	}
}
```



## 3. 消息

###3.1 消息队列

skynet的消息队列实际是全局消息队列加次级消息队列的结构，次级消息队列即各个服务的消息队列，全局队列采用链表实现，次级消息队列采用循环数组实现，容量不够时会扩容为原来的2倍。

skynet_mq.h(skynet-src/skynet_mq.h)中定义了消息结构体和一些方法：

```c
#ifndef SKYNET_MESSAGE_QUEUE_H
#define SKYNET_MESSAGE_QUEUE_H

#include <stdlib.h>
#include <stdint.h>

//消息结构体
struct skynet_message {
	uint32_t source;	//来源
	int session;		//session标识
	void * data;		//消息体
	size_t sz;			//长度
};

// type is encoding in skynet_message.sz high 8bit
#define MESSAGE_TYPE_MASK (SIZE_MAX >> 8)
#define MESSAGE_TYPE_SHIFT ((sizeof(size_t)-1) * 8)

struct message_queue;

//全局消息入队
void skynet_globalmq_push(struct message_queue * queue);
//全局消息出队
struct message_queue * skynet_globalmq_pop(void);

//创建一个消息队列
struct message_queue * skynet_mq_create(uint32_t handle);
void skynet_mq_mark_release(struct message_queue *q);

//消息移除
typedef void (*message_drop)(struct skynet_message *, void *);

//队列释放
void skynet_mq_release(struct message_queue *q, message_drop drop_func, void *ud);
//消息处理者handler
uint32_t skynet_mq_handle(struct message_queue *);

// 0 for success 消息出队
int skynet_mq_pop(struct message_queue *q, struct skynet_message *message);
//消息入队
void skynet_mq_push(struct message_queue *q, struct skynet_message *message);

// return the length of message queue, for debug
int skynet_mq_length(struct message_queue *q);
int skynet_mq_overload(struct message_queue *q);

void skynet_mq_init();

#endif
```

skynet_mq.c(skynet-src/skynet_mq.c)源码：

```c
#include "skynet.h"
#include "skynet_mq.h"
#include "skynet_handle.h"
#include "spinlock.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <assert.h>
#include <stdbool.h>

//默认队列长度为64
#define DEFAULT_QUEUE_SIZE 64
//最大长度为max(16bit) + 1 = 65536 
#define MAX_GLOBAL_MQ 0x10000

// 0 means mq is not in global mq.
// 1 means mq is in global mq , or the message is dispatching.

#define MQ_IN_GLOBAL 1
#define MQ_OVERLOAD 1024

struct message_queue {
	struct spinlock lock;
	uint32_t handle;	//目标handler，唯一标识一个服务
	int cap;			//容量
	int head;			//头位置
	int tail;			//末尾位置
	int release;		//释放标记
	int in_global;		//是否在全局队列中
	int overload;		//最大负载
	int overload_threshold;		//最大负载阈值
	struct skynet_message *queue;	//循环数组
	struct message_queue *next;		//下一个队列，链表
};

//全局消息队列，链表
struct global_queue {
	struct message_queue *head;		//头
	struct message_queue *tail;		//尾
	struct spinlock lock;
};

static struct global_queue *Q = NULL;

void 
skynet_globalmq_push(struct message_queue * queue) {
	struct global_queue *q= Q;

	SPIN_LOCK(q)
	assert(queue->next == NULL);
	if(q->tail) {	//链表不为空，将queue加入到全局消息队列的末尾
		q->tail->next = queue;
		q->tail = queue;
	} else {	//链表为空
		q->head = q->tail = queue;
	}
	SPIN_UNLOCK(q)
}

//取链表中的第一个消息队列
struct message_queue * 
skynet_globalmq_pop() {
	struct global_queue *q = Q;

	SPIN_LOCK(q)
	struct message_queue *mq = q->head;
	if(mq) {
		q->head = mq->next;
		if(q->head == NULL) {
			assert(mq == q->tail);
			q->tail = NULL;
		}
		mq->next = NULL;
	}
	SPIN_UNLOCK(q)

	return mq;
}

//创建一个消息队列
struct message_queue * 
skynet_mq_create(uint32_t handle) {
	struct message_queue *q = skynet_malloc(sizeof(*q));
	q->handle = handle;
	q->cap = DEFAULT_QUEUE_SIZE;
	q->head = 0;	//刚开始头为0
	q->tail = 0;	//刚开始尾为0
	SPIN_INIT(q)
	// When the queue is create (always between service create and service init) ,
	// set in_global flag to avoid push it to global queue .
	// If the service init success, skynet_context_new will call skynet_mq_push to push it to global queue.
	q->in_global = MQ_IN_GLOBAL;
	q->release = 0;
	q->overload = 0;
	q->overload_threshold = MQ_OVERLOAD;
	//这里分配的是数组
	q->queue = skynet_malloc(sizeof(struct skynet_message) * q->cap);
	q->next = NULL;

	return q;
}

//释放队列 回收内存
static void 
_release(struct message_queue *q) {
	assert(q->next == NULL);
	SPIN_DESTROY(q)
	skynet_free(q->queue);
	skynet_free(q);
}

//返回队列的handler
uint32_t 
skynet_mq_handle(struct message_queue *q) {
	return q->handle;
}

//获取队列长度，注意数组被循环使用的情况
int
skynet_mq_length(struct message_queue *q) {
	int head, tail,cap;

	SPIN_LOCK(q)
	head = q->head;
	tail = q->tail;
	cap = q->cap;
	SPIN_UNLOCK(q)
	
	if (head <= tail) {
		return tail - head;
	}
	return tail + cap - head;
}

//获取负载情况
int
skynet_mq_overload(struct message_queue *q) {
	if (q->overload) {
		int overload = q->overload;
		q->overload = 0;	//这里清0是为了避免持续报警，在skynet-server.c中
		return overload;
	} 
	return 0;
}

//消息队列出队
int
skynet_mq_pop(struct message_queue *q, struct skynet_message *message) {
	int ret = 1;
	SPIN_LOCK(q)

	if (q->head != q->tail) {	//消息队列不为空
		*message = q->queue[q->head++];
		ret = 0;
		int head = q->head;
		int tail = q->tail;
		int cap = q->cap;

		if (head >= cap) {	//超出边界，重头开始
			q->head = head = 0;
		}
		int length = tail - head;
		if (length < 0) {
			length += cap;
		}
		//长度要超过阈值了，扩容一倍
		while (length > q->overload_threshold) {
			q->overload = length;
			q->overload_threshold *= 2;
		}
	} else {
		// reset overload_threshold when queue is empty
		q->overload_threshold = MQ_OVERLOAD;
	}

	if (ret) {
		q->in_global = 0;
	}
	
	SPIN_UNLOCK(q)

	return ret;
}

//扩展消息队列
static void
expand_queue(struct message_queue *q) {
	//按原来容量的两倍进行扩容
	struct skynet_message *new_queue = skynet_malloc(sizeof(struct skynet_message) * q->cap * 2);
	int i;
	for (i=0;i<q->cap;i++) {	//拷贝老数据
		new_queue[i] = q->queue[(q->head + i) % q->cap];
	}
	q->head = 0;		//重设head
	q->tail = q->cap;	//重设tail
	q->cap *= 2;		//重设容量
	
	skynet_free(q->queue);	//释放老数组
	q->queue = new_queue;
}

//消息队列入队
void 
skynet_mq_push(struct message_queue *q, struct skynet_message *message) {
	assert(message);
	SPIN_LOCK(q)
	//入队
	q->queue[q->tail] = *message;
	//越界了重头开始
	if (++ q->tail >= q->cap) {
		q->tail = 0;
	}

	//首尾重叠，队列已满，需要扩展
	if (q->head == q->tail) {
		expand_queue(q);
	}

	//重新放回全局队列
	if (q->in_global == 0) {
		q->in_global = MQ_IN_GLOBAL;
		skynet_globalmq_push(q);
	}
	
	SPIN_UNLOCK(q)
}

//初始化全局队列
void 
skynet_mq_init() {
	struct global_queue *q = skynet_malloc(sizeof(*q));
	memset(q,0,sizeof(*q));
	SPIN_INIT(q);
	Q=q;
}

//服务释放标记
void 
skynet_mq_mark_release(struct message_queue *q) {
	SPIN_LOCK(q)
	assert(q->release == 0);
	q->release = 1;
	if (q->in_global != MQ_IN_GLOBAL) {
		skynet_globalmq_push(q);
	}
	SPIN_UNLOCK(q)
}

//释放服务，清空循环数组
static void
_drop_queue(struct message_queue *q, message_drop drop_func, void *ud) {
	struct skynet_message msg;
	while(!skynet_mq_pop(q, &msg)) {
		drop_func(&msg, ud);
	}
	_release(q);	//回收内存
}

//释放服务相关的队列
void 
skynet_mq_release(struct message_queue *q, message_drop drop_func, void *ud) {
	SPIN_LOCK(q)
	
	if (q->release) {
		SPIN_UNLOCK(q)
		_drop_queue(q, drop_func, ud);
	} else {
		skynet_globalmq_push(q);
		SPIN_UNLOCK(q)
	}
}
```

### 3.2 消息处理

本质就是对工作队列中的消息不停地调用回调函数。

skynet是单进程多线程的，线程的种类有monitor/timer/socket/worker，monitor是监控服务是不是陷入死循环了。timer是skynet自己实现的定时器。socket是负责网络的。worker就是工作线程了，monitor/timer/socket都只有一个线程，唯独worker有多个线程，是可配的，不配的话是8个线程。每个工作线程有个叫worker_parm的参数【引用自[skynet源码分析（5）--消息机制之消息处理](https://www.jianshu.com/p/11c46e083a5f)】。

下面是worker的入口函数(skynet-src/skynet_start.c)：

```c
static void *
thread_worker(void *p) {
    struct worker_parm *wp = p;//线程参数
    int id = wp->id; //线程编号
    int weight = wp->weight; //线程权重
    struct monitor *m = wp->m; //monitor，监控器，每个线程有一个
    struct skynet_monitor *sm = m->m[id]; // 线程自己的监控器
    skynet_initthread(THREAD_WORKER);
    struct message_queue * q = NULL;
    while (!m->quit) {
        //消息处理
        q = skynet_context_message_dispatch(sm, q, weight);
        if (q == NULL) { //所有消息队列都是空的
            if (pthread_mutex_lock(&m->mutex) == 0) {
                ++ m->sleep;
                // "spurious wakeup" is harmless,
                // because skynet_context_message_dispatch() can be call at any time.
                if (!m->quit) //不退出的时候等待唤醒
                    pthread_cond_wait(&m->cond, &m->mutex);
                -- m->sleep;
                if (pthread_mutex_unlock(&m->mutex)) {
                    fprintf(stderr, "unlock mutex error");
                    exit(1);
                }
            }
        }
    }
    return NULL;
}
```

该入口函数的功能即不断地执行skynet_context_message_dispatch函数，上skynet_context_message_dispatch的源码(skynet-src/skynet_server.c)：

```c
struct message_queue * 
skynet_context_message_dispatch(struct skynet_monitor *sm, struct message_queue *q, int weight) {
    if (q == NULL) {		//第一次传进来的是null
   		q = skynet_globalmq_pop();		//全局队列出队一个工作队列
   		if (q==NULL)		//q仍然是null 返回null
   			return NULL;
   	}
    
    uint32_t handle = skynet_mq_handle(q);	//获取服务的handle	

	struct skynet_context * ctx = skynet_handle_grab(handle);	//根据handle取出服务的上下文，并将引用计数加1
	if (ctx == NULL) {		//服务被释放了
		struct drop_t d = { handle };
		skynet_mq_release(q, drop_message, &d);		//清空工作队列
		return skynet_globalmq_pop();		//进行下一个工作队列
	}

	int i,n=1;
	struct skynet_message msg;

	for (i=0;i<n;i++) {
		if (skynet_mq_pop(q,&msg)) {		//取工作队列中的消息 skynet_mq_pop返回1则代表为空
			skynet_context_release(ctx);	//工作队列是空的，引用计数减一
			return skynet_globalmq_pop();	//下一个工作队列
		} else if (i==0 && weight >= 0) {	//权重 > 0
			n = skynet_mq_length(q);
			n >>= weight;					//权重越大，循环次数越少
		}
		int overload = skynet_mq_overload(q);
		if (overload) {	//过载
			skynet_error(ctx, "May overload, message queue length = %d", overload);
		}

		//触发monitor
		skynet_monitor_trigger(sm, msg.source , handle);

		if (ctx->cb == NULL) {
			skynet_free(msg.data);
		} else {
			dispatch_message(ctx, &msg);	//消息处理
		}

		//第二次调用，判断是否陷入死循环
		skynet_monitor_trigger(sm, 0,0);
	}

	//把处理机会留给其他服务
	assert(q == ctx->queue);
	struct message_queue *nq = skynet_globalmq_pop();
	if (nq) {
		// If global mq is not empty , push q back, and return next queue (nq)
		// Else (global mq is empty or block, don't push q back, and return q again (for next dispatch)
		skynet_globalmq_push(q);	//把当前队列放回去
		q = nq;						//把机会让给其他队列
	} 
	skynet_context_release(ctx);

	return q;
}
```

**总结：消息的处理就是在worker线程中不断地从全局消息队列中取出一个次级消息队列，对次级消息队列中的消息调用服务注册的回调方法，再将次级消息队列放回全局队列，取下一个次级消息队列。每次处理消息的多少受worker权重限制，权重越大，轮转越快，即每次能处理的消息越少。**



### 3.3 消息的分发

【参考自[skynet源码分析（6）--消息机制之消息分发](https://www.jianshu.com/p/9f37802d478a)】

本质就是消息的入队

消息从服务A发送到服务B，需要经过6层函数。

1. skynet.send(lualib/skynet.lua):

   ```lua
   function skynet.send(addr, typename, ...)
       local p = proto[typename]
       return c.send(addr, p.id, 0 , p.pack(...))
   end
   ```

2. lsend(lualib-src/lua-skynet.c):

   c.send函数位于skynet.core中，而skynet.core是通过c写的供lua调用的库。

   ```lua
   luaopen_skynet_core(lua_State *L) {
       luaL_checkversion(L);
   
      	luaL_Reg l[] = {
      		{ "send" , lsend },
      		{ "genid", lgenid },
      		{ "redirect", lredirect },
           { "command" , lcommand },
       	{ "intcommand", lintcommand },
   		{ "addresscommand", laddresscommand },
   		{ "error", lerror },
   		{ "harbor", lharbor },
   		{ "callback", lcallback },
   		{ "trace", ltrace },
   		{ NULL, NULL },
   	};
   
   	// functions without skynet_context
   	luaL_Reg l2[] = {
   		{ "tostring", ltostring },
   		{ "pack", luaseri_pack },
   		{ "unpack", luaseri_unpack },
   		{ "packstring", lpackstring },
   		{ "trash" , ltrash },
   		{ "now", lnow },
   		{ "hpc", lhpc },	// getHPCounter
   		{ NULL, NULL },
   	};
   
   	lua_createtable(L, 0, sizeof(l)/sizeof(l[0]) + sizeof(l2)/sizeof(l2[0]) -2);
   
   	lua_getfield(L, LUA_REGISTRYINDEX, "skynet_context");
   	struct skynet_context *ctx = lua_touserdata(L,-1);
   	if (ctx == NULL) {
   		return luaL_error(L, "Init skynet context first");
   	}
   
   
   	luaL_setfuncs(L,l,1);    //注册函数
   
   	luaL_setfuncs(L,l2,0);    //注册函数
   
   	return 1;
   }
   
   /*
   	uint32 address
   	 string address
   	integer type
   	integer session
   	string message
   	 lightuserdata message_ptr
   	 integer len
    */
   static int
   lsend(lua_State *L) {
   	return send_message(L, 0, 2);
   }
   ```

   该函数的主要作用为调用send_message函数。

3. send_message(lualib-src/lua-skynet.c):

   ```c
   static int
   send_message(lua_State *L, int source, int idx_type) {
       struct skynet_context * context = lua_touserdata(L, lua_upvalueindex(1));
      	uint32_t dest = (uint32_t)lua_tointeger(L, 1);    //取目标handle
      	const char * dest_string = NULL;
      	if (dest == 0) {
      		if (lua_type(L,1) == LUA_TNUMBER) {
               return luaL_error(L, "Invalid service address 0");
       	}
   		dest_string = get_dest_string(L, 1);
   	}
   
   	int type = luaL_checkinteger(L, idx_type+0);
   	int session = 0;
   	if (lua_isnil(L,idx_type+1)) {
   		type |= PTYPE_TAG_ALLOCSESSION;
   	} else {
   		session = luaL_checkinteger(L,idx_type+1);
   	}
   
   	int mtype = lua_type(L,idx_type+2);
   	switch (mtype) {
   	case LUA_TSTRING: {
   		size_t len = 0;
   		void * msg = (void *)lua_tolstring(L,idx_type+2,&len);
   		if (len == 0) {
   			msg = NULL;
   		}
   		if (dest_string) {
   			session = skynet_sendname(context, source, dest_string, type, session , msg, len);
   		} else {
   			session = skynet_send(context, source, dest, type, session , msg, len);
   		}
   		break;
   	}
   	case LUA_TLIGHTUSERDATA: {
   		void * msg = lua_touserdata(L,idx_type+2);
   		int size = luaL_checkinteger(L,idx_type+3);
   		if (dest_string) {
   			session = skynet_sendname(context, source, dest_string, type | PTYPE_TAG_DONTCOPY, session, msg, size);
   		} else {
   			session = skynet_send(context, source, dest, type | PTYPE_TAG_DONTCOPY, session, msg, size);
   		}
   		break;
   	}
   	default:
   		luaL_error(L, "invalid param %s", lua_typename(L, lua_type(L,idx_type+2)));
   	}
   	if (session < 0) {
   		if (session == -2) {
   			// package is too large
   			lua_pushboolean(L, 0);
   			return 1;
   		}
   		// send to invalid address
   		// todo: maybe throw an error would be better
   		return 0;
   	}
   	lua_pushinteger(L,session);
   	return 1;
   }
   ```

   该函数主要作用为调用skynet_sendname和skynet_send函数，而skynet_sendname的本质也是调用skynet_send函数。

4. skynet_send(skynet-src/skynet_server.c):

   ```c
   int
   skynet_send(struct skynet_context * context, uint32_t source, uint32_t destination , int type, int session, void * data, size_t sz) {
       if ((sz & MESSAGE_TYPE_MASK) != sz) {
      		skynet_error(context, "The message to %x is too large", destination);
      		if (type & PTYPE_TAG_DONTCOPY) {
      			skynet_free(data);
      		}
           return -2;
       }
   	_filter_args(context, type, &session, (void **)&data, &sz);
   
   	if (source == 0) {
   		source = context->handle;
   	}
   
   	if (destination == 0) {
   		if (data) {
   			skynet_error(context, "Destination address can't be 0");
   			skynet_free(data);
   			return -1;
   		}
   
   		return session;
   	}
   	if (skynet_harbor_message_isremote(destination)) {    //分布式
   		struct remote_message * rmsg = skynet_malloc(sizeof(*rmsg));
   		rmsg->destination.handle = destination;
   		rmsg->message = data;
   		rmsg->sz = sz & MESSAGE_TYPE_MASK;
   		rmsg->type = sz >> MESSAGE_TYPE_SHIFT;
   		skynet_harbor_send(rmsg, source, session);
   	} else {    //本地
   		struct skynet_message smsg;
   		smsg.source = source;
   		smsg.session = session;
   		smsg.data = data;
   		smsg.sz = sz;
   
   		if (skynet_context_push(destination, &smsg)) {
   			skynet_free(data);
   			return -1;
   		}
   	}
   	return session;
   }
   ```

   暂时不研究分布式系统，只看本地的情况，主要功能是封装消息体，调用skynet_context_push函数。

5. skynet_context_push(skynet-src/skynet_server.c):

   ```c
   int
   skynet_context_push(uint32_t handle, struct skynet_message *message) {
       struct skynet_context * ctx = skynet_handle_grab(handle);    //给上下文的引用计数加1
      	if (ctx == NULL) {
      		return -1;
      	}
      	skynet_mq_push(ctx->queue, message);    //消息入队
      	skynet_context_release(ctx);    //引用计数减1
   
   	return 0;
   }
   ```

   该函数主要功能是调用skynet_mq_push函数。

6. skynet_mq_push(skynet-src/skynet_mq.c):

   ```c
   //消息队列入队
   void 
   skynet_mq_push(struct message_queue *q, struct skynet_message *message) {
       assert(message);
      	SPIN_LOCK(q)
      	//入队
      	q->queue[q->tail] = *message;
      	//越界了重头开始
       if (++ q->tail >= q->cap) {
   		q->tail = 0;
   	}
   
   	//首尾重叠，需要扩展
   	if (q->head == q->tail) {
   		expand_queue(q);
   	}
   
   	//重新放回全局队列
   	if (q->in_global == 0) {
   		q->in_global = MQ_IN_GLOBAL;
   		skynet_globalmq_push(q);
   	}
   	
   	SPIN_UNLOCK(q)
   }
   ```

   将消息入队，完成消息分发。

   **总结：消息的分发，即服务调用skynet.send()后，消息结构体被加入到该服务对应的次级消息队列中的过程，经过skynet.send()->lsend()->send_message()->skynet_send()->skynet_centext_push()->skynet_mq_push()的过程最终入队。**

### 3.4 消息的回调注册

【参考自[skynet源码分析（10）--消息机制之消息注册和回调](https://www.jianshu.com/p/91fc5db98b37)】

 	消息注册的时候的一般方法是

```lua
local CMD = {}

skynet.dispatch("lua", function(session, source, cmd, ...)
	local f = assert(CMD[cmd])
  	f(...)
end)
```

先看一下dispatch函数的源码(lualib/skunet.lua)：

```lua
function skynet.dispatch(typename, func)
    local p = proto[typename]
   	if func then
   		local ret = p.dispatch
   		p.dispatch = func
   		return ret
    else
   		return p and p.dispatch
   	end
end
```

dispatch函数的作用是将回调函数注册到proto表中对应的消息上。

接下来看服务启动时的源码(lualib/skunet.lua)。

```lua
function skynet.start(start_func)
    c.callback(skynet.dispatch_message)
   	init_thread = skynet.timeout(0, function()
   		skynet.init_service(start_func)
   		init_thread = nil
   	end)
end
```

c.callback函数实际上是在lualib-src/lua-skynet.c中，源码如下：

```c
static int
lcallback(lua_State *L) {
    struct skynet_context * context = lua_touserdata(L, lua_upvalueindex(1));
   	int forward = lua_toboolean(L, 2);
   	luaL_checktype(L,1,LUA_TFUNCTION);
   	lua_settop(L,1);
   	lua_rawsetp(L, LUA_REGISTRYINDEX, _cb);

    lua_rawgeti(L, LUA_REGISTRYINDEX, LUA_RIDX_MAINTHREAD);
	lua_State *gL = lua_tothread(L,-1);

	if (forward) {
		skynet_callback(context, gL, forward_cb);
	} else {
		skynet_callback(context, gL, _cb);
	}

	return 0;
}

//设置回调
void 
skynet_callback(struct skynet_context * context, void *ud, skynet_cb cb) {
    context->cb = cb; 
    context->cb_ud = ud;
}
```

c.callback就是设置回调函数到服务的上下文中，而且设置的是skynet.dispatch_message这个lua方法为回调函数。

从代码中看到，最终调用了skynet_callback这个C函数，这个C函数的第三个参数，是一个中转函数。所以lua服务的回调它不是被直接调的，首先要在_cb这个函数处理一下数据，在_cb里面去调lua的回调函数。_cb这个函数主要就是按照Lua api的协议，将参数准备好，然后调lua的函数。

cb源码(lualib-src/lua-skynet.c):

```c
static int
_cb(struct skynet_context * context, void * ud, int type, int session, uint32_t source, const void * msg, size_t sz) {
	lua_State *L = ud;
	int trace = 1;
	int r;
	int top = lua_gettop(L);
	if (top == 0) {
		lua_pushcfunction(L, traceback);
		lua_rawgetp(L, LUA_REGISTRYINDEX, _cb);
	} else {
		assert(top == 2);
	}
	lua_pushvalue(L,2);		// lua回调函数入栈

	lua_pushinteger(L, type);	//回调参数1，类型
	lua_pushlightuserdata(L, (void *)msg);		//回调参数2，消息体
	lua_pushinteger(L,sz);		//回调参数3，消息长度
	lua_pushinteger(L, session);	//回调参数4，session
	lua_pushinteger(L, source);		//回调参数5，source

	//执行回调函数，即dispatch_message函数，5个参数，0个返回值，trace = 1表示出错调用前面设置的traceback函数
	r = lua_pcall(L, 5, 0 , trace);

	if (r == LUA_OK) {	//执行成功 返回
		return 0;
	}
	const char * self = skynet_command(context, "REG", NULL);
	switch (r) {
	case LUA_ERRRUN:
		skynet_error(context, "lua call [%x to %s : %d msgsz = %d] error : " KRED "%s" KNRM, source , self, session, sz, lua_tostring(L,-1));
		break;
	case LUA_ERRMEM:
		skynet_error(context, "lua memory error : [%x to %s : %d]", source , self, session);
		break;
	case LUA_ERRERR:
		skynet_error(context, "lua error in error : [%x to %s : %d]", source , self, session);
		break;
	case LUA_ERRGCMM:
		skynet_error(context, "lua gc error : [%x to %s : %d]", source , self, session);
		break;
	};

	lua_pop(L,1);

	return 0;
}
```

dispatch_message源码(lualib/skynet.lua):

```lua
function skynet.dispatch_message(...)
    	local succ, err = pcall(raw_dispatch_message,...)
    	while true do
    		local co = tremove(fork_queue,1)
    		if co == nil then
    			break
    		end
    		local fork_succ, fork_err = pcall(suspend,co,coroutine_resume(co))
    		if not fork_succ then
			if succ then
				succ = false
				err = tostring(fork_err)
			else
				err = tostring(err) .. "\n" .. tostring(fork_err)
			end
		end
	end
	assert(succ, tostring(err))
end
```

可以看到dispatch_message实际调用的是raw_dispatch_message函数，源码如下(lualib/skynet.lua):

```lua
local function raw_dispatch_message(prototype, msg, sz, session, source)
    	-- skynet.PTYPE_RESPONSE = 1, read skynet.h
    	if prototype == 1 then
    		local co = session_id_coroutine[session]
		if co == "BREAK" then
			session_id_coroutine[session] = nil
		elseif co == nil then
			unknown_response(session, source, msg, sz)
		else
			local tag = session_coroutine_tracetag[co]
			if tag then c.trace(tag, "resume") end
			session_id_coroutine[session] = nil
			suspend(co, coroutine_resume(co, true, msg, sz))
		end
	else
		local p = proto[prototype]
		if p == nil then
			if prototype == skynet.PTYPE_TRACE then
				-- trace next request
				trace_source[source] = c.tostring(msg,sz)
			elseif session ~= 0 then
				c.send(source, skynet.PTYPE_ERROR, session, "")
			else
				unknown_request(session, source, msg, sz, prototype)
			end
			return
		end

		local f = p.dispatch    //取出真正的回调函数，也就是skynet.dispatch设置的函数
		if f then
			local co = co_create(f)    //创建协程
			session_coroutine_id[co] = session
			session_coroutine_address[co] = source
			local traceflag = p.trace
			if traceflag == false then
				-- force off
				trace_source[source] = nil
				session_coroutine_tracetag[co] = false
			else
				local tag = trace_source[source]
				if tag then
					trace_source[source] = nil
					c.trace(tag, "request")
					session_coroutine_tracetag[co] = tag
				elseif traceflag then
					-- set running_thread for trace
					running_thread = co
					skynet.trace()
				end
			end
			suspend(co, coroutine_resume(co, session,source, p.unpack(msg,sz)))    //唤醒协程
		else
			trace_source[source] = nil
			if session ~= 0 then
				c.send(source, skynet.PTYPE_ERROR, session, "")
			else
				unknown_request(session, source, msg, sz, proto[prototype].name)
			end
		end
	end
end
```

**总结：服务注册回调函数时通过skynet.dispatch()将回调函数注册到proto表中。c层存在一个_cb函数，在lua服务注册时被注册，回调时被调用，回调时取lua中的dispatch_message函数，而dispatch_message通过调用raw_dispatch_message函数找到proto表中消息对应的真正的回调函数放到协程中执行。**



## 4. skynet中的timer

### 4.1 设计思想

skynet的设计思想参考Linux内核动态定时器的机制，【参考[Linux动态内核定时器介绍](// http://www.cnblogs.com/leaven/archive/2010/08/19/1803382.html)】。

```
基于上述考虑，并假定一个定时器要经过interval个时钟滴答后才到期（interval=expires-jiffies）,则Linux采用了下列思想来实现其动态内核定时器机制：对于那些0<=interval<=255的定时器，Linux严格按照定时器向量的基本语义来阻止这些定时器，也即Linux内核最关心那些在接下来255个时钟节拍内就要到期的定时器，因此将他们按照各自不通的expires值组织成256个定时器向量。而对于那些256<=interval<=0xffffffff的定时器，由于他们离到期还有一段时间，因此内核并不关心他们，而是将他们以一种扩展的定时器向量语义（或称为“松散的定时器向量语义”）进行组织。所谓“松散的定时器向量语义”就是指：各定时器的expires值可以互不相同的一个定时器队列。
```

在skynet里，时间精度是10ms，这对于游戏服务器来说已经足够了，定义1滴答=0.01秒，1秒=100滴答。其核心思想是：每个定时器设置一个到期的滴答数，与当前系统的滴答数（启动时是0，然后1滴答1滴答往后跳）比较差值，如果差值interval比较小（[0,255]），表示定时器即将到来，需要严格关注，把他们保存在2^8个定时器链表里；如果interval越大，表示定时器越远，可以不用太关注，划分成4个等级，[2^8,2^(8+6)-1]，[2^(8+6),2^(8+6+6)-1]，[2^(8+6+6),2^(8+6+6+6)-1]，[2^(8+6+6+6),2^(8+6+6+6+6)-1]，每个等级只需要2^6个定时器链表保存，比如对于[2^8,2^(8+6)-1]的定时器，将interval>>8相同的值idx保存在第一个等级位置为idx的链表里。

这样做的优势是：不用为每一个interval创建一个链表，而只需要2^8+4*(2^6)个链表，大大节省了内存。

skynet中的timer的源码在skynet-src/skynet_timer.c和skynet-src/skynet_timer.h中。

###4.2 源码

skynet-src/skynet_start.c中timer线程源码：

```c
static void *
thread_timer(void *p) {
	struct monitor * m = p;
	skynet_initthread(THREAD_TIMER);
	for (;;) {
        //每2.5ms调用一次skynet_updatetime()
		skynet_updatetime();
		skynet_socket_updatetime();
		CHECK_ABORT
		wakeup(m,m->count-1);
		usleep(2500);
		if (SIG) {
			signal_hup();
			SIG = 0;
		}
	}
	// wakeup socket thread
	skynet_socket_exit();
	// wakeup all worker thread
	pthread_mutex_lock(&m->mutex);
	m->quit = 1;
	pthread_cond_broadcast(&m->cond);
	pthread_mutex_unlock(&m->mutex);
	return NULL;
}
```

skynet_timer.h源码:

```c
#ifndef SKYNET_TIMER_H
#define SKYNET_TIMER_H

#include <stdint.h>

int skynet_timeout(uint32_t handle, int time, int session);
void skynet_updatetime(void);
uint32_t skynet_starttime(void);
uint64_t skynet_thread_time(void);	// for profile, in micro second

void skynet_timer_init(void);

#endif

```

skyner_time.c源码:

```c
#include "skynet.h"

#include "skynet_timer.h"
#include "skynet_mq.h"
#include "skynet_server.h"
#include "skynet_handle.h"
#include "spinlock.h"

#include <time.h>
#include <assert.h>
#include <string.h>
#include <stdlib.h>
#include <stdint.h>

#if defined(__APPLE__)
#include <AvailabilityMacros.h>
#include <sys/time.h>
#include <mach/task.h>
#include <mach/mach.h>
#endif

typedef void (*timer_execute_func)(void *ud,void *arg);

#define TIME_NEAR_SHIFT 8
#define TIME_NEAR (1 << TIME_NEAR_SHIFT)
#define TIME_LEVEL_SHIFT 6
#define TIME_LEVEL (1 << TIME_LEVEL_SHIFT)
#define TIME_NEAR_MASK (TIME_NEAR-1)
#define TIME_LEVEL_MASK (TIME_LEVEL-1)

struct timer_event {
	uint32_t handle;
	int session;
};

struct timer_node {
	struct timer_node *next;
	uint32_t expire;
};

struct link_list {
	struct timer_node head;
	struct timer_node *tail;
};

//timer结构体
struct timer {
	struct link_list near[TIME_NEAR];		//临近的定时器数组，2^8个链表
	struct link_list t[4][TIME_LEVEL];		//四个级别的定时器数组，4*(2^6)个链表
	struct spinlock lock;
	uint32_t time;				//计数器，程序从启动到现在的滴答数
	uint32_t starttime;			//程序启动的时间点，秒级
	uint64_t current;			//从程序启动到现在的耗时，精度10毫秒级
	uint64_t current_point;		//当前时间，精度10毫秒级
};

static struct timer * TI = NULL;

//清空list链表，并返回第一个节点
static inline struct timer_node *
link_clear(struct link_list *list) {
	struct timer_node * ret = list->head.next;
	list->head.next = 0;
	list->tail = &(list->head);

	return ret;
}

//将节点插入链表尾部
static inline void
link(struct link_list *list,struct timer_node *node) {
	list->tail->next = node;
	list->tail = node;
	node->next=0;
}

//添加一个定时器节点
static void
add_node(struct timer *T,struct timer_node *node) {
	uint32_t time=node->expire;
	uint32_t current_time=T->time;
	
	if ((time|TIME_NEAR_MASK)==(current_time|TIME_NEAR_MASK)) {
        //定时器到期滴答数跟当前比较接近（<2^8）
		link(&T->near[time&TIME_NEAR_MASK],node);
	} else {
        //定时器距离过期还有一段时间，添加到对应的T->t[i]中
		int i;
		uint32_t mask=TIME_NEAR << TIME_LEVEL_SHIFT;
		for (i=0;i<3;i++) {
			if ((time|(mask-1))==(current_time|(mask-1))) {
				break;
			}
			mask <<= TIME_LEVEL_SHIFT;
		}

		link(&T->t[i][((time>>(TIME_NEAR_SHIFT + i*TIME_LEVEL_SHIFT)) & TIME_LEVEL_MASK)],node);	
	}
}

static void
timer_add(struct timer *T,void *arg,size_t sz,int time) {
	struct timer_node *node = (struct timer_node *)skynet_malloc(sizeof(*node)+sz);
	memcpy(node+1,arg,sz);

	SPIN_LOCK(T);

		node->expire=time+T->time;
		add_node(T,node);

	SPIN_UNLOCK(T);
}

//移动某个级别的链表内容
static void
move_list(struct timer *T, int level, int idx) {
	struct timer_node *current = link_clear(&T->t[level][idx]);
	while (current) {
		struct timer_node *temp=current->next;
		add_node(T,current);
		current=temp;
	}
}

//定时器的移动
static void
timer_shift(struct timer *T) {
	int mask = TIME_NEAR;
	uint32_t ct = ++T->time;
	if (ct == 0) {
		move_list(T, 3, 0);
	} else {
		uint32_t time = ct >> TIME_NEAR_SHIFT;
		int i=0;

		while ((ct & (mask-1))==0) {
			int idx=time & TIME_LEVEL_MASK;
			if (idx!=0) {
				move_list(T, i, idx);
				break;				
			}
			mask <<= TIME_LEVEL_SHIFT;
			time >>= TIME_LEVEL_SHIFT;
			++i;
		}
	}
}

static inline void
dispatch_list(struct timer_node *current) {
	do {
        //发送消息
		struct timer_event * event = (struct timer_event *)(current+1);
		struct skynet_message message;
		message.source = 0;
		message.session = event->session;
		message.data = NULL;
		message.sz = (size_t)PTYPE_RESPONSE << MESSAGE_TYPE_SHIFT;

		skynet_context_push(event->handle, &message);
		
		struct timer_node * temp = current;
		current=current->next;
		skynet_free(temp);	
	} while (current);
}

static inline void
timer_execute(struct timer *T) {
	int idx = T->time & TIME_NEAR_MASK;
	
	while (T->near[idx].head.next) {
		struct timer_node *current = link_clear(&T->near[idx]);
		SPIN_UNLOCK(T);
		// dispatch_list don't need lock T
		dispatch_list(current);
		SPIN_LOCK(T);
	}
}

static void 
timer_update(struct timer *T) {
	SPIN_LOCK(T);

	// try to dispatch timeout 0 (rare condition)
	timer_execute(T);

	// shift time first, and then dispatch timer message
	timer_shift(T);

	timer_execute(T);

	SPIN_UNLOCK(T);
}

static struct timer *
timer_create_timer() {
	struct timer *r=(struct timer *)skynet_malloc(sizeof(struct timer));
	memset(r,0,sizeof(*r));

	int i,j;

	for (i=0;i<TIME_NEAR;i++) {
		link_clear(&r->near[i]);
	}

	for (i=0;i<4;i++) {
		for (j=0;j<TIME_LEVEL;j++) {
			link_clear(&r->t[i][j]);
		}
	}

	SPIN_INIT(r)

	r->current = 0;

	return r;
}


int
skynet_timeout(uint32_t handle, int time, int session) {
	if (time <= 0) {
		struct skynet_message message;
		message.source = 0;
		message.session = session;
		message.data = NULL;
		message.sz = (size_t)PTYPE_RESPONSE << MESSAGE_TYPE_SHIFT;

		if (skynet_context_push(handle, &message)) {
			return -1;
		}
	} else {	//创建一个定时器
		struct timer_event event;
		event.handle = handle;
		event.session = session;
		timer_add(TI, &event, sizeof(event), time);
	}

	return session;
}

// centisecond: 1/100 second
static void
systime(uint32_t *sec, uint32_t *cs) {
#if !defined(__APPLE__) || defined(AVAILABLE_MAC_OS_X_VERSION_10_12_AND_LATER)
	struct timespec ti;
	clock_gettime(CLOCK_REALTIME, &ti);
	*sec = (uint32_t)ti.tv_sec;
	*cs = (uint32_t)(ti.tv_nsec / 10000000);
#else
	struct timeval tv;
	gettimeofday(&tv, NULL);
	*sec = tv.tv_sec;
	*cs = tv.tv_usec / 10000;
#endif
}

static uint64_t
gettime() {
	uint64_t t;
#if !defined(__APPLE__) || defined(AVAILABLE_MAC_OS_X_VERSION_10_12_AND_LATER)
	struct timespec ti;
	clock_gettime(CLOCK_MONOTONIC, &ti);
	t = (uint64_t)ti.tv_sec * 100;
	t += ti.tv_nsec / 10000000;
#else
	struct timeval tv;
	gettimeofday(&tv, NULL);
	t = (uint64_t)tv.tv_sec * 100;
	t += tv.tv_usec / 10000;
#endif
	return t;
}

//每2.5毫秒触发一次
void
skynet_updatetime(void) {
	uint64_t cp = gettime();
	if(cp < TI->current_point) {
		skynet_error(NULL, "time diff error: change from %lld to %lld", cp, TI->current_point);
		TI->current_point = cp;
	} else if (cp != TI->current_point) {
		uint32_t diff = (uint32_t)(cp - TI->current_point);
		TI->current_point = cp;    //当前时间，毫秒级
		TI->current += diff;		//从启动到现在耗时
		int i;
		for (i=0;i<diff;i++) {
			timer_update(TI);
		}
	}
}

uint32_t
skynet_starttime(void) {
	return TI->starttime;
}

uint64_t 
skynet_now(void) {
	return TI->current;
}

void 
skynet_timer_init(void) {
	TI = timer_create_timer();
	uint32_t current = 0;
	systime(&TI->starttime, &current);
	TI->current = current;
	TI->current_point = gettime();
}

// for profile

#define NANOSEC 1000000000
#define MICROSEC 1000000

uint64_t
skynet_thread_time(void) {
#if  !defined(__APPLE__) || defined(AVAILABLE_MAC_OS_X_VERSION_10_12_AND_LATER)
	struct timespec ti;
	clock_gettime(CLOCK_THREAD_CPUTIME_ID, &ti);

	return (uint64_t)ti.tv_sec * MICROSEC + (uint64_t)ti.tv_nsec / (NANOSEC / MICROSEC);
#else
	struct task_thread_times_info aTaskInfo;
	mach_msg_type_number_t aTaskInfoCount = TASK_THREAD_TIMES_INFO_COUNT;
	if (KERN_SUCCESS != task_info(mach_task_self(), TASK_THREAD_TIMES_INFO, (task_info_t )&aTaskInfo, &aTaskInfoCount)) {
		return 0;
	}

	return (uint64_t)(aTaskInfo.user_time.seconds) + (uint64_t)aTaskInfo.user_time.microseconds;
#endif
}

```

###4.3 总结

timer的工作流程是timer线程每2.5ms调用一次skynet_updatetime()方法，skynet_updatetime()会调用time_update()方法，该方法除了触发定时器外，还需要重新分配定时器所在区间（timer_shift）。

因为T->near里保存即将触发的定时器，所以每TIME_NEAR-1（2^8-1）个滴答数才有可能需要分配（第22行）。否则，分配T->t中某个等级即可。

当T->time的低8位不全为0时，不需要分配，所以每2^8个滴答数才有需要分配一次；

当T->time的第9-14位不全为0时，重新分配T[0]等级，每2^8个滴答数分配一次，idx从1开始，每次分配+1；

当T->time的第15-20位不全为0时，重新分配T[1]等级，每2^(8+6)个滴答数分配一次，idx从1开始，每次分配+1；

当T->time的第21-26位不全为0时，重新分配T[2]等级，每2^(8+6+6)个滴答数分配一次，idx从1开始，每次分配+1；

当T->time的第27-32位不全为0时，重新分配T[3]等级，每2^(8+6+6+6)个滴答数分配一次，idx从1开始，每次分配+1；

即等级越大的定时器越遥远，越不关注，需要重新分配的次数也就越少。



## 5. 网络

[参考自[skynet源码分析（8）--skynet的网络](https://www.jianshu.com/p/364ea070557f)、[skynet源码分析之网络层——Lua层介绍](https://www.cnblogs.com/RainRill/p/8707328.html)】

### 5.1 socket线程

之前提到skynet的线程种类中有一类socket线程，这个线程的代码如下：

```c
static void *
thread_socket(void *p) {
    struct monitor * m = p;
    skynet_initthread(THREAD_SOCKET);
    for (;;) {
        int r = skynet_socket_poll();
        if (r==0)
            break;
        if (r<0) {
            CHECK_ABORT
            continue;
        }
        wakeup(m,0);
    }
    return NULL;
}
```

看skynet_socket_poll部分的源码

```c
int 
skynet_socket_poll() {
	struct socket_server *ss = SOCKET_SERVER;	//并没有被重新初始化
	assert(ss);
	struct socket_message result;
	int more = 1;
	int type = socket_server_poll(ss, &result, &more);
	switch (type) {
	case SOCKET_EXIT:
		return 0;
	case SOCKET_DATA:
		forward_message(SKYNET_SOCKET_TYPE_DATA, false, &result);
		break;
	case SOCKET_CLOSE:
		forward_message(SKYNET_SOCKET_TYPE_CLOSE, false, &result);
		break;
	case SOCKET_OPEN:
		forward_message(SKYNET_SOCKET_TYPE_CONNECT, true, &result);
		break;
	case SOCKET_ERR:
		forward_message(SKYNET_SOCKET_TYPE_ERROR, true, &result);
		break;
	case SOCKET_ACCEPT:
		forward_message(SKYNET_SOCKET_TYPE_ACCEPT, true, &result);
		break;
	case SOCKET_UDP:
		forward_message(SKYNET_SOCKET_TYPE_UDP, false, &result);
		break;
	case SOCKET_WARNING:
		forward_message(SKYNET_SOCKET_TYPE_WARNING, false, &result);
		break;
	default:
		skynet_error(NULL, "Unknown socket message type %d.",type);
		return -1;
	}
	if (more) {
		return -1;
	}
	return 1;
}
```

其中会调用socket_server_poll函数，看一下socket_server_poll的源码：

```c
// return type
int 
socket_server_poll(struct socket_server *ss, struct socket_message * result, int * more) {
	for (;;) {
		if (ss->checkctrl) {
			if (has_cmd(ss)) {
				int type = ctrl_cmd(ss, result);
				if (type != -1) {
					clear_closed_event(ss, result, type);
					return type;
				} else
					continue;
			} else {
				ss->checkctrl = 0;
			}
		}
		if (ss->event_index == ss->event_n) {
			ss->event_n = sp_wait(ss->event_fd, ss->ev, MAX_EVENT);
			ss->checkctrl = 1;
			if (more) {
				*more = 0;
			}
			ss->event_index = 0;
			if (ss->event_n <= 0) {
				ss->event_n = 0;
				if (errno == EINTR) {
					continue;
				}
				return -1;
			}
		}
		struct event *e = &ss->ev[ss->event_index++];
		struct socket *s = e->s;
		if (s == NULL) {
			// dispatch pipe message at beginning
			continue;
		}
		struct socket_lock l;
		socket_lock_init(s, &l);
		switch (s->type) {
		case SOCKET_TYPE_CONNECTING:
			return report_connect(ss, s, &l, result);
		case SOCKET_TYPE_LISTEN: {
			int ok = report_accept(ss, s, result);
			if (ok > 0) {
				return SOCKET_ACCEPT;
			} if (ok < 0 ) {
				return SOCKET_ERR;
			}
			// when ok == 0, retry
			break;
		}
		case SOCKET_TYPE_INVALID:
			fprintf(stderr, "socket-server: invalid socket\n");
			break;
		default:
			if (e->read) {
				int type;
				if (s->protocol == PROTOCOL_TCP) {
					type = forward_message_tcp(ss, s, &l, result);
				} else {
					type = forward_message_udp(ss, s, &l, result);
					if (type == SOCKET_UDP) {
						// try read again
						--ss->event_index;
						return SOCKET_UDP;
					}
				}
				if (e->write && type != SOCKET_CLOSE && type != SOCKET_ERR) {
					// Try to dispatch write message next step if write flag set.
					e->read = false;
					--ss->event_index;
				}
				if (type == -1)
					break;				
				return type;
			}
			if (e->write) {
				int type = send_buffer(ss, s, &l, result);
				if (type == -1)
					break;
				return type;
			}
			if (e->error) {
				// close when error
				int error;
				socklen_t len = sizeof(error);  
				int code = getsockopt(s->fd, SOL_SOCKET, SO_ERROR, &error, &len);  
				const char * err = NULL;
				if (code < 0) {
					err = strerror(errno);
				} else if (error != 0) {
					err = strerror(error);
				} else {
					err = "Unknown error";
				}
				force_close(ss, s, &l, result);
				result->data = (char *)err;
				return SOCKET_ERR;
			}
			if(e->eof) {
				force_close(ss, s, &l, result);
				return SOCKET_CLOSE;
			}
			break;
		}
	}
}
```

再追踪ctrl_cmd（skynet-src/socket_server.c）的源码：

```c
// return type
static int
ctrl_cmd(struct socket_server *ss, struct socket_message *result) {
	int fd = ss->recvctrl_fd;
	// the length of message is one byte, so 256+8 buffer size is enough.
	uint8_t buffer[256];
	uint8_t header[2];
	block_readpipe(fd, header, sizeof(header));
	int type = header[0];
	int len = header[1];
	block_readpipe(fd, buffer, len);
	// ctrl command only exist in local fd, so don't worry about endian.
	switch (type) {
	case 'S':
		return start_socket(ss,(struct request_start *)buffer, result);
	case 'B':
		return bind_socket(ss,(struct request_bind *)buffer, result);
	case 'L':
		return listen_socket(ss,(struct request_listen *)buffer, result);
	case 'K':
		return close_socket(ss,(struct request_close *)buffer, result);
	case 'O':
		return open_socket(ss, (struct request_open *)buffer, result);
	case 'X':
		result->opaque = 0;
		result->id = 0;
		result->ud = 0;
		result->data = NULL;
		return SOCKET_EXIT;
	case 'D':
	case 'P': {
		int priority = (type == 'D') ? PRIORITY_HIGH : PRIORITY_LOW;
		struct request_send * request = (struct request_send *) buffer;
		int ret = send_socket(ss, request, result, priority, NULL);
		dec_sending_ref(ss, request->id);
		return ret;
	}
	case 'A': {
		struct request_send_udp * rsu = (struct request_send_udp *)buffer;
		return send_socket(ss, &rsu->send, result, PRIORITY_HIGH, rsu->address);
	}
	case 'C':
		return set_udp_address(ss, (struct request_setudp *)buffer, result);
	case 'T':
		setopt_socket(ss, (struct request_setopt *)buffer);
		return -1;
	case 'U':
		add_udp_socket(ss, (struct request_udp *)buffer);
		return -1;
	default:
		fprintf(stderr, "socket-server: Unknown ctrl %c.\n",type);
		return -1;
	};

	return -1;
}
```

socket线程一直调用skynet_socket_poll函数，这和普通网络服务器的写法是一样的。普通网络服务器也是创建socket，绑定socket，添加到epoll，然后epoll_wait等待事件的发生。当socket上发生相应的事件后，根据事件类型forward_message函数向消息队列中添加消息。worker线程在之后的某个时刻处理消息。

### 5.2 socket的连接过程

先看一段简单的示例代码

```lua
skynet.start(function()
    local fd = socket.listen("127.0.0.1",8888)
    socket.start(fd, function (fd, addr)
    	socket.start(fd)
    	...
    end)

    server = skynet.newservice("server")
end)
```



skynet中的socket结构有几种状态：

```c
#define SOCKET_TYPE_INVALID 0 //可使用
#define SOCKET_TYPE_RESERVE 1 //已占用
#define SOCKET_TYPE_PLISTEN 2 //等待监听(监听套接字拥有)
#define SOCKET_TYPE_LISTEN 3 //监听，可接受客户端的连接（监听套接字才拥有）
#define SOCKET_TYPE_CONNECTING 4 //正在连接（connect失败时状态，tcp会尝试重新connect）
#define SOCKET_TYPE_CONNECTED 5 //已连接，可以收发数据
#define SOCKET_TYPE_HALFCLOSE 6
#define SOCKET_TYPE_PACCEPT 7 //等待连接（连接套接字才拥有）
#define SOCKET_TYPE_BIND 8
```

socket.listen（lualib/skynet/socket.lua）源码：

```lua
function socket.listen(host, port, backlog)
	if port == nil then
		host, port = string.match(host, "([^:]+):(.+)$")
		port = tonumber(port)
	end
	return driver.listen(host, port, backlog)
end
```

socket.listen调用了driver.listen，追踪driver.listen（lualib-src/lua-socket.c）源码：

```c
static int
llisten(lua_State *L) {
	const char * host = luaL_checkstring(L,1);
	int port = luaL_checkinteger(L,2);
	int backlog = luaL_optinteger(L,3,BACKLOG);
	struct skynet_context * ctx = lua_touserdata(L, lua_upvalueindex(1));
	int id = skynet_socket_listen(ctx, host,port,backlog);
	if (id < 0) {
		return luaL_error(L, "Listen error");
	}

	lua_pushinteger(L,id);
	return 1;
}
```

再追踪skynet_socket_listen（skynet-src/skynet_socket.c）源码：

```c
int 
skynet_socket_listen(struct skynet_context *ctx, const char *host, int port, int backlog) {
	uint32_t source = skynet_context_handle(ctx);
	return socket_server_listen(SOCKET_SERVER, source, host, port, backlog);
}
```

继续追踪socket_server_listen（skynet-src/socket_server.c）源码：

```c
int 
socket_server_listen(struct socket_server *ss, uintptr_t opaque, const char * addr, int port, int backlog) {
	int fd = do_listen(addr, port, backlog);
	if (fd < 0) {
		return -1;
	}
	struct request_package request;
	int id = reserve_id(ss);
	if (id < 0) {
		close(fd);
		return id;
	}
	request.u.listen.opaque = opaque;
	request.u.listen.id = id;
	request.u.listen.fd = fd;
	send_request(ss, &request, 'L', sizeof(request.u.listen));
	return id;
}
```

socket_server_listen函数会调用send_request（skynet-src/socket_server.c）函数，源码：

```c
static void
send_request(struct socket_server *ss, struct request_package *request, char type, int len) {
	request->header[6] = (uint8_t)type;
	request->header[7] = (uint8_t)len;
	for (;;) {
		ssize_t n = write(ss->sendctrl_fd, &request->header[6], len+2);
		if (n<0) {
			if (errno != EINTR) {
				fprintf(stderr, "socket-server : send ctrl command error %s.\n", strerror(errno));
			}
			continue;
		}
		assert(n == len+2);
		return;
	}
}
```

到这个函数以后，比较清楚地看到，数据被发送到sendctrl_fd这个描述符上了。而ctrl_cmd在recvctrl_fd描述符上接收数据，因为sendctrl_fd和recvctrl_fd是一个管道的发送端和接收端。随后，ctrl_cmd中执行listen_socket函数，追踪listen_socket（skynet-src/socket_server.c）的源码：

```c
static int
listen_socket(struct socket_server *ss, struct request_listen * request, struct socket_message *result) {
	int id = request->id;
	int listen_fd = request->fd;
	struct socket *s = new_fd(ss, id, listen_fd, PROTOCOL_TCP, request->opaque, false);
	if (s == NULL) {
		goto _failed;
	}
	s->type = SOCKET_TYPE_PLISTEN;
	return -1;
_failed:
	close(listen_fd);
	result->opaque = request->opaque;
	result->id = id;
	result->ud = 0;
	result->data = "reach skynet socket number limit";
	ss->slot[HASH_ID(id)].type = SOCKET_TYPE_INVALID;

	return SOCKET_ERR;
}

```

执行完listen_socket函数后socket的type变为SOCKET_TYPE_PLISTEN。

随后开始执行socket.start(fd, function (fd, addr)...)，追踪socket.start（）的源码：

```lua
function socket.start(id, func)
	driver.start(id)
	return connect(id, func)
end
```

再通过driver.start追踪到lstart（lualib-src/lua-socket.c）函数，源码：

```C
static int
lstart(lua_State *L) {
	struct skynet_context * ctx = lua_touserdata(L, lua_upvalueindex(1));
	int id = luaL_checkinteger(L, 1);
	skynet_socket_start(ctx,id);
	return 0;
}
```

该函数又调用了skynet_socket_start（skynet-src/skynet_socket.c），源码：

```c
void 
skynet_socket_start(struct skynet_context *ctx, int id) {
	uint32_t source = skynet_context_handle(ctx);
	socket_server_start(SOCKET_SERVER, source, id);
}
```

该函数又调用了socket_server_start（）函数，源码：

```c
void 
socket_server_start(struct socket_server *ss, uintptr_t opaque, int id) {
	struct request_package request;
	request.u.start.id = id;
	request.u.start.opaque = opaque;
	send_request(ss, &request, 'S', sizeof(request.u.start));
}
```

该函数发送一条消息，追踪到ctrl_cmd中，调用了start_socket函数，源码如下：

```C
static int
start_socket(struct socket_server *ss, struct request_start *request, struct socket_message *result) {
	int id = request->id;
	result->id = id;
	result->opaque = request->opaque;
	result->ud = 0;
	result->data = NULL;
	struct socket *s = &ss->slot[HASH_ID(id)];
	if (s->type == SOCKET_TYPE_INVALID || s->id !=id) {
		result->data = "invalid socket";
		return SOCKET_ERR;
	}
	struct socket_lock l;
	socket_lock_init(s, &l);
	if (s->type == SOCKET_TYPE_PACCEPT || s->type == SOCKET_TYPE_PLISTEN) {
		if (sp_add(ss->event_fd, s->fd, s)) {
			force_close(ss, s, &l, result);
			result->data = strerror(errno);
			return SOCKET_ERR;
		}
		s->type = (s->type == SOCKET_TYPE_PACCEPT) ? SOCKET_TYPE_CONNECTED : SOCKET_TYPE_LISTEN;	
		s->opaque = request->opaque;
		result->data = "start";
		return SOCKET_OPEN;
	} else if (s->type == SOCKET_TYPE_CONNECTED) {
		// todo: maybe we should send a message SOCKET_TRANSFER to s->opaque
		s->opaque = request->opaque;
		result->data = "transfer";
		return SOCKET_OPEN;
	}
	// if s->type == SOCKET_TYPE_HALFCLOSE , SOCKET_CLOSE message will send later
	return -1;
}
```

此时socket的type为SOCKET_TYPE_PLISTEN，所以s->type = SOCKET_TYPE_LISTEN，此时socket可以等待客户端的连接请求了。

当客户端发起连接请求后，epoll事件返回，调用report_accept（skynet-src/socket_server.c）方法，源码如下：

```c
static int
report_accept(struct socket_server *ss, struct socket *s, struct socket_message *result) {
	union sockaddr_all u;
	socklen_t len = sizeof(u);
	int client_fd = accept(s->fd, &u.s, &len);
	if (client_fd < 0) {
		if (errno == EMFILE || errno == ENFILE) {
			result->opaque = s->opaque;
			result->id = s->id;
			result->ud = 0;
			result->data = strerror(errno);
			return -1;
		} else {
			return 0;
		}
	}
	int id = reserve_id(ss);
	if (id < 0) {
		close(client_fd);
		return 0;
	}
	socket_keepalive(client_fd);
	sp_nonblocking(client_fd);
	struct socket *ns = new_fd(ss, id, client_fd, PROTOCOL_TCP, s->opaque, false);
	if (ns == NULL) {
		close(client_fd);
		return 0;
	}
	// accept new one connection
	stat_read(ss,s,1);

	ns->type = SOCKET_TYPE_PACCEPT;		//注意这里
	result->opaque = s->opaque;
	result->id = s->id;
	result->ud = id;
	result->data = NULL;

	if (getname(&u, ss->buffer, sizeof(ss->buffer))) {
		result->data = ss->buffer;
	}

	return 1;
}
```

此时，ns->type被设置为SOCKET_TYPE_PACCEPT。

随后触发回调函数，执行socket.start(fd)，此时start_socket中会将socket的状态置为`SOCKET_TYPE_CONNECTED`。

```C
s->type = (s->type == SOCKET_TYPE_PACCEPT) ? SOCKET_TYPE_CONNECTED : SOCKET_TYPE_LISTEN;
```

至此，连接建立完成，可以收发数据。