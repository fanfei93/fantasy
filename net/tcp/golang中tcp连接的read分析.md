## 1 tcp socket缓冲区已满是否会丢包

不会。

tcp发送数据时滑动窗口协议维护两个窗口结构，发送窗口结构和接收窗口结构，发送窗口中有接收方提供的表示接收缓存区剩余大小（即当前能接受的数据大小），所以当接受方缓冲区满了之后，发送方得知接收方无法接收数据后会阻塞。

具体参看[TCP传输控制协议(3)--数据传输(滑动窗口)](https://img-blog.csdnimg.cn/20190115200355848.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L0pvY2tlcl9E,size_16,color_FFFFFF,t_70)



## 2 golang中`io.ReadFull`和`net.TCPConn.Read`

先看用例代码：

```go
// conn *net.TCPConn
// headData []byte
_, e := io.ReadFull(conn, headData)
```

再看io.ReadFull的源码

```go

// ReadFull reads exactly len(buf) bytes from r into buf.
// It returns the number of bytes copied and an error if fewer bytes were read.
// The error is EOF only if no bytes were read.
// If an EOF happens after reading some but not all the bytes,
// ReadFull returns ErrUnexpectedEOF.
// On return, n == len(buf) if and only if err == nil.
// If r returns an error having read at least len(buf) bytes, the error is dropped.
func ReadFull(r Reader, buf []byte) (n int, err error) {
	return ReadAtLeast(r, buf, len(buf))
}



// ReadAtLeast reads from r into buf until it has read at least min bytes.
// It returns the number of bytes copied and an error if fewer bytes were read.
// The error is EOF only if no bytes were read.
// If an EOF happens after reading fewer than min bytes,
// ReadAtLeast returns ErrUnexpectedEOF.
// If min is greater than the length of buf, ReadAtLeast returns ErrShortBuffer.
// On return, n >= min if and only if err == nil.
// If r returns an error having read at least min bytes, the error is dropped.
func ReadAtLeast(r Reader, buf []byte, min int) (n int, err error) {
	if len(buf) < min {
		return 0, ErrShortBuffer
	}
	for n < min && err == nil {
		var nn int
		nn, err = r.Read(buf[n:])	//看这里
		n += nn
	}
	if n >= min {
		err = nil
	} else if n > 0 && err == EOF {
		err = ErrUnexpectedEOF
	}
	return
}
```

通过上面的源码可以看出`io.ReadFull`的本质是调用了`r.Read`，只是在没有读到指定长度的数据时不跳出循环而继续调用`r.Read`。

而`r.Read`在这里即`net.TCPConn.Read`，定位到源码

```go

// Read implements the Conn Read method.
func (c *conn) Read(b []byte) (int, error) {
	if !c.ok() {
		return 0, syscall.EINVAL
	}
	n, err := c.fd.Read(b)		//看这里
	if err != nil && err != io.EOF {
		err = &OpError{Op: "read", Net: c.fd.net, Source: c.fd.laddr, Addr: c.fd.raddr, Err: err}
	}
	return n, err
}


// c.fd.Read()
func (fd *netFD) Read(p []byte) (n int, err error) {
	n, err = fd.pfd.Read(p)		//看这里
	runtime.KeepAlive(fd)
	return n, wrapSyscallError("read", err)
}

// fd.pfd.Read()
// Read implements io.Reader.
func (fd *FD) Read(p []byte) (int, error) {
	if err := fd.readLock(); err != nil {
		return 0, err
	}
	defer fd.readUnlock()
	if len(p) == 0 {
		// If the caller wanted a zero byte read, return immediately
		// without trying (but after acquiring the readLock).
		// Otherwise syscall.Read returns 0, nil which looks like
		// io.EOF.
		// TODO(bradfitz): make it wait for readability? (Issue 15735)
		return 0, nil
	}
	if err := fd.pd.prepareRead(fd.isFile); err != nil {
		return 0, err
	}
	if fd.IsStream && len(p) > maxRW {
		p = p[:maxRW]
	}
	for {
		n, err := syscall.Read(fd.Sysfd, p)		//看这里
		if err != nil {
			n = 0
			if err == syscall.EAGAIN && fd.pd.pollable() {
				if err = fd.pd.waitRead(fd.isFile); err == nil {
					continue
				}
			}

			// On MacOS we can see EINTR here if the user
			// pressed ^Z.  See issue #22838.
			if runtime.GOOS == "darwin" && err == syscall.EINTR {
				continue
			}
		}
		err = fd.eofError(n, err)
		return n, err
	}
}

// syscall.Read()
func Read(fd int, p []byte) (n int, err error) {
	n, err = read(fd, p)	//看这里
	if race.Enabled {
		if n > 0 {
			race.WriteRange(unsafe.Pointer(&p[0]), n)
		}
		if err == nil {
			race.Acquire(unsafe.Pointer(&ioSync))
		}
	}
	if msanenabled && n > 0 {
		msanWrite(unsafe.Pointer(&p[0]), n)
	}
	return
}

// read()
func read(fd int, p []byte) (n int, err error) {
	var _p0 unsafe.Pointer
	if len(p) > 0 {
		_p0 = unsafe.Pointer(&p[0])
	} else {
		_p0 = unsafe.Pointer(&_zero)
	}
	r0, _, e1 := syscall(funcPC(libc_read_trampoline), uintptr(fd), uintptr(_p0), uintptr(len(p)))		//看这里
	n = int(r0)
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

func libc_read_trampoline()

//go:linkname libc_read libc_read
//go:cgo_import_dynamic libc_read read "/usr/lib/libSystem.B.dylib"
```

最后定位到`r0, _, e1 := syscall(funcPC(libc_read_trampoline), uintptr(fd), uintptr(_p0), uintptr(len(p)))`这行代码，这行代码其实是调用一个系统函数去执行读取操作。

首先根据`go:linkname libc_read libc_read` 得知`libc_read_trampoline()`实际指向`libc_read`函数。

再根据`go:cgo_import_dynamic libc_read read "/usr/lib/libSystem.B.dylib"` 得知`libc_read`函数实际指向的是动态库`/usr/lib/libSystem.B.dylib`中的read函数。

最终执行的是系统提供的`read`函数

```c
#include <unistd.h>
ssize_t read(int fd, void *buf, size_t count);
```

