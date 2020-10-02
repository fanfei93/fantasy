参考[skynet设计综述](https://blog.codingnow.com/2012/09/the_design_of_skynet.html)

## skynet概述

### skynet核心解决什么问题

把一个符合规范的C模块，从动态库（so文件）中启动起来，绑定一个永不重复（即使模块退出）的数字id做为其handle。模块被称为服务，服务间可以自由发送消息。每个模块可以向skynet框架注册一个callback函数，用来接收发送给它的消息。每个服务都是被一个个消息包驱动，当没有包到来的时候，它们就处于挂起状态，对cpu资源零消耗。

### skynet核心不解决什么问题

和普通的单线程程序一样，你要为你代码中的bug和意外负责，如果你的程序出了问题而崩溃，你不应该把错误藏起来，假装它们没有发生。

简单说，skynet只负责把一个数据包从一个服务内发送出去，让同一进程内的另一个服务收到，调用对应的callback函数处理。它保证，模块的初始化过程，每个独立的callback调用，都是相互线程安全的。

### skynet的消息调度

skynet维护了两级消息队列。

每个服务实体有一个私有的消息队列，队列中是一个个发送给它的消息。消息由四部分构成：

```c
struct skynet_message {
    uint32_t source;
    int session;
    void * data;
    size_t sz;
};
```

向一个服务发送一个消息，就是把这样一个消息体压入这个服务的私有消息队列中。这个结构的值复制进消息队列的，但消息内容本身不做复制。

skynet维护了一个全局消息队列，里面放的是诸个不为空的次级消息队列。

在skynet启动时，建立了若干工作线程，它们不断的从主消息队列中取出一个次级消息队列来，再从次级队列中取去一条消息，调用对应的服务的callback函数进行处理。为了调用公平，一次仅处理一条消息，而不是消耗所有消息（虽然那样的局部效率更高，因为减少了查询服务实体的次数，以及主消息队列进出的次数），这样可以保证没有服务被饿死。

用户定义的callback函数不必保证线程安全，因为在callback函数被调用的过程中，其它工作线程没有可能获得这个callback函数锁属服务的次级消息队列，也就不可能被并发了。一旦一个服务的消息队列暂时为空，它的消息队列就不再被放回全局消息队列了，这样使大部分不工作的服务不会空转cpu。

### Gate和Connection

一个完整的游戏服务器不可避免地要和外界通讯。

外界通讯有两种，一是游戏客户端使用TCP连接接入skynet节点。

另一个是第三方的服务，比如数据库服务，它接收一个或多个TCP连接，你需要从skynet内部建立一个TCP连接出去使用。

前者称为gate服务，它的特征是监听一个TCP端口，接受连入的TCP连接，并把连接上获得的数据转发到skynet内部。gate可以用来消除外部数据包和skynet内部消息包的不一致性。外部TCP流的分包问题，是gate实现上的约定。

gate会接受外部连接，并把连接相关信息转发给另一个服务去处理。它自己不做数据处理是因为我们需要保持gate实现的简洁高效。

外部信息分为两类，一类是连接本身的接入和断开消息，另一类是连接上的数据包。一开始，gate无条件转发这两类消息到同一个处理服务，但对于连接数据包，添加一个包头无疑有性能上的开销。所以gate还接收另一种工作模式：把每个不同连接上的数据包转发给不通的独立服务上。每个独立服务处理单一连接上的数据包。

或者，我们也可以选择把不同连接上的数据包从公控制信息包中分离开，但不区分不同连接而转发给同一数据处理服务（对数据来源不敏感，对数据内容敏感的场合）。

这三种模式，分别称为

+ watchdog模式，由gate加上包头，同时处理控制信息和数据信息的所有数据
+ agent模式，让每个agent处理独立连接
+ broker模式，由一个broker服务处理不同连接上的所有数据包

无论哪种模式，控制信息都是交由watchdog去处理的。

注意，gate只负责读取外部数据，但不负责回写。也就是说，向这些连接发送数据不是它的职责范畴。

另一个重要组件叫Connection。它和Gate不通，它负责从skynet内部建立socket到外部服务。

Connection分两个部分，一部分用于监听不同的系统fd的可读状态，这是用epoll实现的。它收到这个连接上的数据后，会把所有数据不做任何分包，转发到另一个服务里去处理。这和gate的行为不太一致，这是因为connection多用于使用外部第三方数据库，我们很难统一其分包的格式。

另一部分是Lua相关的底层支持库，可以用于建立连接，以及连接上数据常用的分包规则。



### lua服务的启动过程

如果要创建一个lua服务，需要使用skynet.newservice。假设现在在A服务，我们需要创建B服务，这个流程是这样的：

+ A服务调用skynet.newservice(name,...)，这个函数使A阻塞。
+ B被创建出来，name.lua这个脚本被执行，脚本要调用skynet.start(function...end)，表示服务B启动，可以接受消息。
+ 当上面skynet.start的函数返回时，A的skynet.newservice才返回，并且A得到了B的服务句柄。



## Lua API

每个skynet服务，最重要的职业就是处理别的服务发送过来的消息，以及向别的服务发送消息。每条skynet消息由五个元素构成。

+ session：大部分消息工作在请求回应模式下，即，一个服务向另一个服务发起一个请求，而后收到请求的服务在处理完请求消息后，回复一条消息。session是由发起请求的服务生成的，对它自己唯一的消息标识。回应方在回应时，将session带回。这样发送方才能识别出哪条消息是针对哪条的回应。session是一个非负整数，当一条消息不需要回应时，按惯例，使用0这个特殊的session号。session由skynet框架生成管理，通常不需要使用者关心。
+ source：消息源。每个服务都由一个32bit整数标识。这个整数可以看成是服务在skynet系统中的地址，即使在服务退出后，新启动的服务通常也不会使用已用过的地址（除非发生回绕，但一般间隔时间非常长）。每条收到的消息都携带有source，方便在回应的时候可以指定地址。但地址的管理通常由框架完成，用户不用关心。
+ type：消息类别。每个服务可以接收256种不同类别的消息。每种类别可以有不同的消息编码格式。有十几种类别是框架保留的，通常也不建议用户定义新的消息类别。因为用户完全可以利用已有的类别，而用具体的消息内容区分每条具体的含义。框架把这些type映射为字符串便于记忆。最常用的消息类别名为"lua"广泛用于用lua编写的skynet服务间的通讯。
+ message：消息的C指针，在lua层看来是一个lightuserdata。框架会隐藏这个细节，最终用户处理的是经过解码过的lua对象，只有极少情况，你才需要在lua层直接操作这个指针。
+ size：消息的长度。通常和message一起结合起来使用。



## 服务地址

每个服务都有一个32bit的数字地址，这个地址的高8bit表明了它所属的节点。

`skynet.self()`用于获得服务自己的地址。

`skynet.harbor()`用于获得服务器所属的节点。

`skynet.address(address)`用于把一个地址数字转换为一个可用于阅读的字符串。

同时我们还可以给地址起一个名字方便使用。

`skynet.register(name)`可以为自己注册一个别名。（别名必须在16字符以内）

`skynet.name(name,address)`为一个地址命名。`skynet.name(name,skynet.self())`和`skynet.register(name)`功能等价。

这个名字一旦注册，是在skynet系统中通用的，你需要自己约定名字的管理的方法。

以`.`开头的名字是在统一skynet节点下有效的，跨节点的skynet服务对别的节点下的`.`开头的名字不可见。不同的skynet节点可以定义相同的`.`开头的名字。

以字母开头的名字在整个skynet网络中都有效，你可以通过这种全局名字把消息发到其他节点。原则上，不鼓励滥用全局名字，它有一定的管理成本。管用的方法是在业务层交换服务的数字地址，让服务自行记住其他服务的地址来传播消息。

`skynet.localname(name)`用来查询一个`.`开头的名字对应的地址。它是一个非阻塞api，不可以查询跨节点的全局名字。

## 消息分发和回应

`skynet.dispatch(type,function(session,source,...) ... end)`注册特定类消息的处理函数。大多数程序会注册lua类消息的处理函数，惯例的写法是：

```lua
local CMD = {}

skynet.dispatch("lua", function(session, source, cmd, ...)
  local f = assert(CMD[cmd])
  f(...)
end)
```

这段代码注册了lua类消息的分发函数。通常约定lua类消息的第一个元素是一个字符串，表示具体消息对应的操作。我们会在脚本中创建一个CMD表，把对应的操作函数定义在表中。每条lua消息抵达后，从CMD表中查到处理函数，并把余下的参数传入。这个消息的session和source可以不必传递给处理函数，因为除了主动向source发送类别为response的消息来回应它以为，还有更简单的方法。框架记忆了这两个值。

虽然并不推荐，但你还可以注册新的消息类别，方法是使用`skynet.register_protocol`。例如你可以注册一个以文本方式编码消息的消息类别。通常由C编写的服务更容易解析文本消息。skynet已经定义了这种消息类别为`skynet.PTYPE_TEXT`，但默认没有注册到lua中使用。

```lua
skynet.register_protocol {
	name = "text",
	id = skynet.PTYPE_TEXT,
	pack = function(m) return tostring(m) end,
	unpack = skynet.tostring,
}
```

新的类别必须提供pack和unpack函数，用于消息的编码和解码。

pack函数必须返回一个string或是一个userdata和size。在lua脚本中，推荐你返回string类型，而用后一种形式需要对skynet底层有足够的了解（采用它多半是因为性能考虑，可以减少一些数据拷贝）。

unpack函数接收一个lightuserdata和一个整数。即上面提到的message和size。lua无法直接处理c指针，所以必须使用额外的C库导入函数来解码。skynet.tostring就是这样一个函数，它将这个C指针和长度翻译成lua的string。

接下来你可以使用`skynet.dispatch`注册text类别的处理方法了。当然，直接在`skynet.register_protocol`时传入dispatch函数也可以。

dispatch函数会在收到每条类别对应的消息时被回调。消息先经过unpack函数，返回值被传入dispatch。每条消息的处理都工作在一个独立的coroutine中，看起来以多线程方式工作。但记住，在同一个lua虚拟机（同一个lua服务中），永远不可能出现多线程并发的情况。你的lua脚本不需要考虑线程安全的问题，但每次有阻塞api调用时，脚本都可能发生重入，这点务必小心。

回应一个消息可以使用`skynet.ret(message,size)`。它会将message size对应的消息附上当前消息的session，以及`skynet.PTYPE_RESPONSE`这个类别，发送给当前消息的来源source。由于某些历史原因（早起的skynet默认消息类别是文本，而没有经过特殊编码），这个API被设计成传递一个C指针和长度，而不是经过当前消息的pack函数打包。或者你也可以省略size而传入一个字符串。

由于skynet中最常用的消息类别是lua，这种消息是经过`skynet.pack`打包的，所以惯用法是`skynet.ret(skynet.pack(...))`。btw，`skynet.pack(...)`返回一个lightuserdata和一个长度，符合skynet.ret参数要求，与之对应的是`skynet.unpack(message,size)`它可以把一个C指针加长度的消息解码成一组lua对象。

skynet.ret在同一个消息处理的coroutine中只可以被调用一次，多次调用会触发异常。有时候，你需要挂起一个请求，等将来实际满足，再回应它。而回应的时候已经在别的coroutine中了。针对这种情况，你可以调用		`skynet.response(skynt.pack)`获得一个闭包，以后调用这个闭包即可把回应消息发回。这里的参数skynet.pack是可选的，你可以传入其他打包函数，默认即是skynet.pack。

`skynet.response`返回的闭包可用于延迟回应。调用它时，第一个参数通常是true表示是一个正常的回应，之后的参数是需要回应的数据。如果是false，则给请求者抛出一个异常。它的返回值表示回应的地址是否还有效。如果你仅仅想知道回应地址的有效性，那么可以在第一个参数传入“TEST”用于检测。

注：`skynet.ret`和`skynet.response`都是非阻塞api。

如果你没有回应（ret或response）一个外部请求，session不为0时，skynet会写一条log提醒你这里可能有问题。你可以调用`skynet.ignoreret()`告诉框架你打算忽略这个session，这通常在你想利用session传递其它数据时（即，不用`skynet.call`调用）使用。比如，你可以将客户端的socket fd当成session，把外部消息直接转发给内部服务处理。

### 关于消息数据指针

skynet服务间传递的消息在底层是用C指针/lightuserdata加一个数字长度来表示的。当一条消息进入skynet服务时，该消息会根据消息类别分发到对应的类别处理流程，（由`skynet.register_protocol`）。这个消息数据指针是由发送消息方生成的，通常是由skynet_malloc分配的内存块。默认情况下，框架会在之后调用`skynet_free`释放这个指针。

如果你想阻止框架调用`skynet_free`可以使用`skynet.forward_type`取代`skynet.start`调用。和`skynet.start`不同，`skynet_forwardtype`需要多传递一张表，表示哪些类的消息不需要框架调用`skynet_free`。例如：

```lua
skynet.forward_type({[skynet.PTYPE_LUA] = skynet.PTYPE_USER},start_func)
```

表示PTYPE_LUA类的消息处理完毕后，不要调用`skynet_free`释放消息数据指针。这通常用于做消息转发。

这是由于框架默认定义了`PTYPE_LUA`的处理流程，而`skynet.register_protocol`不准重定义这个流程，所以我们可以重定向消息类型为`PTYPE_USER`。

还有另外一种情况也需要用`skynet.forward_type`阻止释放消息数据指针：如果针对某种特别的消息，传了一个复杂对象（而不是由skynet_malloc分配出来的整块内存）那么就可以让框架忽略数据指针，而自己调用对象的释放函数去释放这个指针。



## 消息的序列化

当我们能确保消息仅在同一进程间流通的时候，便可以直接把C对象编码成一个指针。因为进程相同，所以C指针可以有效传递。但是，skynet默认支持有多节点模式，消息有可能被传到另一台机器的另一个进程中。这种情况下，每条消息都必须是一块连续内存，我们就必须对消息进行序列化操作。

skynet默认提供了一套对lua数据结构的序列化方案。即上一节提到的`skynet.pack`以及`skynet.unpack`函数。`skynet.pack`可以将一组lua对象序列化为一个由malloc分配出来的C指针加一个数字长度。你需要考虑C指针引用的数据块何时释放的问题。当然，如果你只是将`skynet.pack`填在消息处理框架里时，框架解决了这个管理问题。skynet将C指针发送到其他服务，而接收方会在使用完后释放这个指针。

如果你想把这个序列化模块做它用，建议使用另一个api `skynet.packstring`。和`skynet.pack`不通，它返回一个lua string。而`skynet.unpack`即可以处理C指针，也可以处理lua string。

这个序列化库支持string，boolean，number，lightuserdata，table这些类型，但对lua table的metatable支持非常有限，所以尽量不要用其打包带有元方法的lua对象。



## 消息推送和远程调用

`skynet.send(address,typename,...)`这条api可以把一条类别为typename的消息发送给address。它会先经过事先注册的pack函数打包`...`的内容。

`skynet.send`是一条非阻塞api，发送完消息后，coroutine会继续向下运行，这期间服务不会重入。

`skynet.call(address,typename,...)`这条api则不同，它会在内部生成一个唯一session，并向address提起请求，并阻塞等待对session的回应（可以不由address回应）。当消息回应后，还会通过之前的注册的unpack函数解包。表面上看起来，就是发起了一次rpc，并阻塞等待回应。call不支持超时。

尤其需要留意的是，`skynet.call`仅仅阻塞住当前的coroutine，而没有阻塞整个服务。在等待回应期间，服务照样可以响应其他请求。所以尤其要注意，在skynet.call之前获得的服务内的状态，到返回后，很有可能改变。



## 服务的启动和退出

每个skynet服务都必须有一个启动函数。这一点和普通lua脚本不同，传统的lua脚本是没有专门的主函数，脚本本身即是主函数。而skynet服务，你必须主动调用`skynet.start(function() ... end)`。

`skynet.start`注册一个函数为这个服务的启动函数。当然你还是可以在脚本中随意写一段lua代码，他们会先于start函数执行。但是，不要在外面调用skynet的阻塞api，因为框架将无法唤醒它们。

如果你想在`skynet.start`注册的函数之前做点什么，可以调用`skynet.init(function() ... end)`。这通常用于lua库的编写。你需要编写的服务引用你的库的时候，事先调用一些skynet阻塞api，就可以用`skynet.init`把这些工作注册在start之前。

`skynet.exit()`用于退出当前的服务。`skynet.exit`之后的代码都不会被执行。而且，当前服务被阻塞住的coroutine也会立刻中断退出。这些通常是一些RPC尚未收到回应。所以调用`skynet.exit()`请务必小心。

`skynet.kill(address)`可以用来强制关闭别的服务。但强烈不推荐这样做。因为对象会在任意一条消息处理完毕后，毫无征兆的退出。所以推荐的做法是，发送一条消息，让对方善后以及调用`skynet.exit`。注：`skynet.kill(skynet.self())`不完全等价于`skynet.exit()`，后者更安全。

`skynet.newservice(name,...)`用于启动一个新的lua服务。name是脚本的名字（不用写.lua后缀）。只有被启动的脚本的start函数返回后，这个API才会返回启动的服务的地址，这是一个阻塞的API。如果被启动的脚本在初始化环节抛出异常，或在初始化完成前就调用`skynet.exit`退出，`skynet.newservice`都会抛出异常。如果被启动的脚本的start函数是一个用不结束的循环，那么newserivice也会被永远阻塞住。





