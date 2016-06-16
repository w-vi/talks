// This file is a complementary document for the talk on golang
// channels internals I gave on [golang meetup in
// Brno](http://www.meetup.com/Golang-Brno/events/231266490/)(Czech
// Republic) on 16th of June 2016.  Process it with
// [gocco](http://nikhilm.github.io/gocco/) to render html out of it.

// Part of the talk was quick walk-through of the go channel code in
// `go/src/runtime/chan.go` file and this file is a transcript of
// sorts of that walk-through not a very detailed one but enough to get
// the idea. I am keeping some of the original comments in place as they help
// to understand the code.

package runtime

// This file contains the full implementation of Go channels but has
// been changed. Well just the comments the go code is intact. Please
// see
// [`go/src/runtime/chan.go`](https://github.com/golang/go/blob/7e460e70d90295cf08ea627c0a0fff170aba5518/src/runtime/chan.go)
// for the original content.
import (
	"runtime/internal/atomic"
	"unsafe"
)

// Aligning memory to 8 bytes, that's why the cryptic `hchanSize`
const (
	maxAlign  = 8
	hchanSize = unsafe.Sizeof(hchan{}) + uintptr(-int(unsafe.Sizeof(hchan{}))&(maxAlign-1))
	debugChan = false
)

// ## The Channel
// `hchan struct` is the representation of the channel itself.  It is
// implemented as a ring buffer (circular queue) with given size
// `dataqsiz`. The memory for the buffer is stored in `buf` pointer
// and is allocated dynamically based on the type of the channel. It
// stores also a lists of waiting senders and receivers in `sendq` and
// `recvq` respectively. `elemsize` is the size of elements
// stored. See the comments int the source for some more details.

type hchan struct {
	qcount   uint           // total data in the queue
	dataqsiz uint           // size of the circular queue
	buf      unsafe.Pointer // points to an array of dataqsiz elements
	elemsize uint16
	closed   uint32
	elemtype *_type // element type
	sendx    uint   // send index
	recvx    uint   // receive index
	recvq    waitq  // list of recv waiters
	sendq    waitq  // list of send waiters

	// *Original Comment:*
	//
	// *lock protects all fields in hchan, as well as several
	// fields in sudogs blocked on this channel.*
	//
	// *Do not change another G's status while holding this lock
	// (in particular, do not ready a G), as this can deadlock
	// with stack shrinking.*
	lock mutex
}

// ## Invariants
// Through out there are some invariants when it come to channels.
// 1) At least one of c.sendq and c.recvq is empty.
// *For buffered channels, also:*
// 2) c.qcount > 0 implies that c.recvq is empty.
// 3) c.qcount < c.dataqsiz implies that c.sendq is empty.

// `waitq struct` is a linked list of `sudog` structs. sudog represents a goroutine in a channel wait list, see the definition of [`type sudog`](https://github.com/golang/go/blob/7fdec6216c0a25c6dbcc8159b755da6682dd9080/src/runtime/runtime2.go#L235) in the github source of golang.
type waitq struct {
	first *sudog
	last  *sudog
}

func reflect_makechan(t *chantype, size int64) *hchan {
	return makechan(t, size)
}

// ## Creating channel
//`makechan` creates the channel as such, it is a direct translation
//of `make(chan type)`, `size` can be zero. The `chantype` represents
//the channel as a
//[type](https://github.com/golang/go/blob/797dc584577c66ee1e181a3f423133ee83647247/src/runtime/type.go#L346)
//contains information like what elements are transferred.
func makechan(t *chantype, size int64) *hchan {
	elem := t.elem

	// This block should not happen but just to make sure :)
	if elem.size >= 1<<16 {
		throw("makechan: invalid channel element type")
	}
	if hchanSize%maxAlign != 0 || elem.align > maxAlign {
		throw("makechan: bad alignment")
	}
	if size < 0 || int64(uintptr(size)) != size || (elem.size > 0 && uintptr(size) > (_MaxMem-hchanSize)/elem.size) {
		panic(plainError("makechan: size out of range"))
	}

	// First comes the allocation of the `hchan.buf` in one go for unbuffered channel or channel with elements of no size (`make(chan struct{})`).
	var c *hchan
	if elem.kind&kindNoPointers != 0 || size == 0 {
		// *Original Comments:*
		//
		// *Allocate memory in one call.
		// Hchan does not contain pointers interesting for GC in this case:
		// buf points into the same allocation, elemtype is persistent.
		// SudoG's are referenced from their owning thread so they can't be collected.*
		c = (*hchan)(mallocgc(hchanSize+uintptr(size)*elem.size, nil, true))
		if size > 0 && elem.size != 0 {
			// `add` makes the old fashioned pointer arithmetics, base + offset
			c.buf = add(unsafe.Pointer(c), hchanSize)
		} else {
			c.buf = unsafe.Pointer(c)
		}
	} else {
		// Allocates the `hchan` struct first and then the ring buffer,
		// buffered channels with non-zero sized elements.
		c = new(hchan)
		c.buf = newarray(elem, int(size))
	}
	c.elemsize = uint16(elem.size)
	c.elemtype = elem
	c.dataqsiz = uint(size)

	if debugChan {
		print("makechan: chan=", c, "; elemsize=", elem.size, "; elemalg=", elem.alg, "; dataqsiz=", size, "\n")
	}
	return c
}

// <a name="chanbuf"/> `chanbuf(c, i)` is a helper function which
// returns pointer to the i'th slot in the channels ring buffer.
func chanbuf(c *hchan, i uint) unsafe.Pointer {
	return add(c.buf, uintptr(i)*uintptr(c.elemsize))
}

//## Send
// entry point for `c <- x` from compiled code
func chansend1(t *chantype, c *hchan, elem unsafe.Pointer) {
	chansend(t, c, elem, true, getcallerpc(unsafe.Pointer(&t)))
}

//<a name = "chansend" />
// `chansend` is the implementation of the `c <- x`

//- `ep` is the pointer to the element in question (on stack of sending goroutine)
//- `block` comes from the `select` statement, should the sender block or not? If not then the goroutine will not sleep but return if it could not complete
//- `callerpc` is callers program counter, where the caller is.
//
// The return value is `bool` indicating if the send happened or not
// (useful for `select` again)
func chansend(t *chantype, c *hchan, ep unsafe.Pointer, block bool, callerpc uintptr) bool {
	if raceenabled {
		raceReadObjectPC(t.elem, ep, callerpc, funcPC(chansend))
	}
	if msanenabled {
		msanread(ep, t.elem.size)
	}
	// **AXIOM #1** - A send to a `nil` channel blocks forever
	if c == nil {
		if !block {
			return false
		}
		// `gopark` puts the calling goroutine to sleep queue
		gopark(nil, nil, "chan send (nil chan)", traceEvGoStop, 2)
		throw("unreachable")
	}

	if debugChan {
		print("chansend: chan=", c, "\n")
	}

	if raceenabled {
		racereadpc(unsafe.Pointer(c), callerpc, funcPC(chansend))
	}

	// *Original Comments:
	// Fast path: check for failed non-blocking operation without acquiring the lock.*
	//
	// *After observing that the channel is not closed, we observe that the channel is
	// not ready for sending. Each of these observations is a single word-sized read
	// (first c.closed and second c.recvq.first or c.qcount depending on kind of channel).
	// Because a closed channel cannot transition from 'ready for sending' to
	// 'not ready for sending', even if the channel is closed between the two observations,
	// they imply a moment between the two when the channel was both not yet closed
	// and not ready for sending. We behave as if we observed the channel at that moment,
	// and report that the send cannot proceed.*
	//
	// *It is okay if the reads are reordered here: if we observe that the channel is not
	// ready for sending and then observe that it is not closed, that implies that the
	// channel wasn't closed during the first observation.*
	//
	// Fast check whether there is a chance to send.
	//
	// If not blocking and the channel is not closed and one of the
	// following conditions are met bail out and return false, we
	// can't send at the moment.
	//
	// 1. buffer has zero size and there is no receiver (unbuffered channel)
	// 2. buffer have size and is full (full buffered channel)
	if !block && c.closed == 0 && ((c.dataqsiz == 0 && c.recvq.first == nil) ||
		(c.dataqsiz > 0 && c.qcount == c.dataqsiz)) {
		return false
	}

	// We acquire the `lock` so nobody is going to change the state
	// under our hands.
	var t0 int64
	if blockprofilerate > 0 {
		t0 = cputicks()
	}

	lock(&c.lock)

	//<a name="axiom2"/>
	// **AXIOM #2** - A send to a closed channel panics
	if c.closed != 0 {
		unlock(&c.lock)
		panic(plainError("send on closed channel"))
	}

	if sg := c.recvq.dequeue(); sg != nil {
		// Found a waiting receiver. We pass the value we want to send
		// directly to the receiver, bypassing the channel buffer (if any).
		// [`send`](#send)
		send(c, sg, ep, func() { unlock(&c.lock) })
		return true
	}

	if c.qcount < c.dataqsiz {
		// Space is available in the channel buffer. Enqueue the element to send.
		// `qp` points to `sendx` index in the `hchan.buf` ring buffer.
		//[`chanbuf`](#chanbuf)
		qp := chanbuf(c, c.sendx)
		if raceenabled {
			raceacquire(qp)
			racerelease(qp)
		}
		// Move the element to the ring buffer.
		typedmemmove(c.elemtype, qp, ep)
		c.sendx++
		// Because it is a ring buffer we wrap the `sendx` if it is
		// pointing beyond the buffer size.
		if c.sendx == c.dataqsiz {
			c.sendx = 0
		}
		c.qcount++
		// Unlock and return true as the value was sent.
		unlock(&c.lock)
		return true
	}

	// We are at the point where we couldn't store the value in the
	// ring buffer as it is full, so if it should not block bail out now.
	if !block {
		unlock(&c.lock)
		return false
	}

	// Block on the channel. Some receiver will complete our operation
	// for us.  Create new `sudog` struct for current sender and
	// channel and enqueue it to the wait linked list.
	// gp is a `g` struct representing goroutine see the [definition](https://github.com/golang/go/blob/7fdec6216c0a25c6dbcc8159b755da6682dd9080/src/runtime/runtime2.go#L306)
	gp := getg()
	mysg := acquireSudog()
	mysg.releasetime = 0
	if t0 != 0 {
		mysg.releasetime = -1
	}
	mysg.elem = ep
	mysg.waitlink = nil
	mysg.g = gp
	mysg.selectdone = nil
	mysg.c = c
	gp.waiting = mysg
	gp.param = nil
	c.sendq.enqueue(mysg)
	// Here the goroutine goes to sleep and scheduler will wake it
	// once there is a receiver on this channel. The lock is released
	// inside the function.
	goparkunlock(&c.lock, "chan send", traceEvGoBlockSend, 3)

	// This comes after the blocking, the goroutine is awaken.  We
	// don't store the data anymore, a receiving goroutine
	// handled the transfer already.
	if mysg != gp.waiting {
		throw("G waiting list is corrupted")
	}
	gp.waiting = nil
	if gp.param == nil {
		// Goroutines `param` can be `nil` only if the channel was
		// closed. Spurious wakeups should not happen
		if c.closed == 0 {
			throw("chansend: spurious wakeup")
		}
		panic(plainError("send on closed channel"))
	}

	// Release everything and return, the value was sent.
	gp.param = nil
	if mysg.releasetime > 0 {
		blockevent(mysg.releasetime-t0, 2)
	}
	mysg.c = nil
	releaseSudog(mysg)
	return true
}

// <a name="send"></a>
// `send` processes a send operation on an empty channel c.
// The value `ep` sent by the sender is copied to the receiver `sg`.
// The receiver is then woken up to go on its merry way.
//
// *Conditions:*
//
// - Channel c must be empty and locked.
// - Send unlocks it with `unlockf` function.
// - `sg` must already be dequeued from c.
// - `ep` must be non-nil and point to the heap or the caller's stack.
func send(c *hchan, sg *sudog, ep unsafe.Pointer, unlockf func()) {
	if raceenabled {
		if c.dataqsiz == 0 {
			racesync(c, sg)
		} else {
			qp := chanbuf(c, c.recvx)
			raceacquire(qp)
			racerelease(qp)
			raceacquireg(sg.g, qp)
			racereleaseg(sg.g, qp)
			c.recvx++
			if c.recvx == c.dataqsiz {
				c.recvx = 0
			}
			c.sendx = c.recvx // c.sendx = (c.sendx+1) % c.dataqsiz
		}
	}
	//`sg.elem` is a pointer to the channel element type and holds the sent
	//value so must not be nil. ['sendDirect'](#sendDirect) is where it happens.
	if sg.elem != nil {
		sendDirect(c.elemtype, sg, ep)
		sg.elem = nil
	}
	gp := sg.g
	unlockf()
	// Set the receiver `g.param` so it is non-nil to signal that the channel was not closed. Used after the [wakeup](#rcv_wakeup).
	gp.param = unsafe.Pointer(sg)
	if sg.releasetime != 0 {
		sg.releasetime = cputicks()
	}
	// Wake the receiving goroutine
	goready(gp, 4)
}

//<a name ="sendDirect"/>
// `sendDirect` is where the transfer happens. See the original comments.
func sendDirect(t *_type, sg *sudog, src unsafe.Pointer) {
	// *Original Comments:*
	//
	// *Send on an unbuffered or empty-buffered channel is the only operation
	// in the entire runtime where one goroutine
	// writes to the stack of another goroutine. The GC assumes that
	// stack writes only happen when the goroutine is running and are
	// only done by that goroutine. Using a write barrier is sufficient to
	// make up for violating that assumption, but the write barrier has to work.
	// typedmemmove will call heapBitsBulkBarrier, but the target bytes
	// are not in the heap, so that will not help. We arrange to call
	// memmove and typeBitsBulkBarrier instead.*

	// *Once we read sg.elem out of sg, it will no longer
	// be updated if the destination's stack gets copied (shrunk).
	// So make sure that no preemption points can happen between read & use.*
	dst := sg.elem
	memmove(dst, src, t.size)
	typeBitsBulkBarrier(t, uintptr(dst), t.size)
}

// ## Close
// Direct translation of `close(chan)`
func closechan(c *hchan) {
	// Closing non existing channel panics!
	if c == nil {
		panic(plainError("close of nil channel"))
	}

	// Lock the channel so we don't mess other goroutines.
	lock(&c.lock)
	// Double closing panics. Closing on channel is the last signal
	// saying there will be no more, doing the same thing again
	// violates it.
	if c.closed != 0 {
		unlock(&c.lock)
		panic(plainError("close of closed channel"))
	}

	if raceenabled {
		callerpc := getcallerpc(unsafe.Pointer(&c))
		racewritepc(unsafe.Pointer(c), callerpc, funcPC(closechan))
		racerelease(unsafe.Pointer(c))
	}
	// The channel is closed now.
	c.closed = 1

	var glist *g

	// Release all readers on the queue.
	// Sets the receivers `g.param` to `nil` so it is clear that the channel
	// was closed.
	for {
		sg := c.recvq.dequeue()
		if sg == nil {
			break
		}
		if sg.elem != nil {
			memclr(sg.elem, uintptr(c.elemsize))
			sg.elem = nil
		}
		if sg.releasetime != 0 {
			sg.releasetime = cputicks()
		}
		gp := sg.g
		gp.param = nil
		if raceenabled {
			raceacquireg(gp, unsafe.Pointer(c))
		}
		gp.schedlink.set(glist)
		glist = gp
	}

	// Release all waiting senders.
	// Sets the senders `g.param` to `nil` so it is clear that the channel
	// was closed and they will panic! [AXIOM #2](#axiom2)
	for {
		sg := c.sendq.dequeue()
		if sg == nil {
			break
		}
		sg.elem = nil
		if sg.releasetime != 0 {
			sg.releasetime = cputicks()
		}
		gp := sg.g
		gp.param = nil
		if raceenabled {
			raceacquireg(gp, unsafe.Pointer(c))
		}
		gp.schedlink.set(glist)
		glist = gp
	}
	unlock(&c.lock)

	// Now schedule all the released goroutines, the channel lock was dropped.
	for glist != nil {
		gp := glist
		glist = glist.schedlink.ptr()
		gp.schedlink = 0
		goready(gp, 3)
	}
}

// ## Receive
// entry points for `<- c` from compiled code
func chanrecv1(t *chantype, c *hchan, elem unsafe.Pointer) {
	chanrecv(t, c, elem, true)
}

func chanrecv2(t *chantype, c *hchan, elem unsafe.Pointer) (received bool) {
	_, received = chanrecv(t, c, elem, true)
	return
}

// `chanrecv` receives on channel c and writes the received data to `ep`. It is almost the same as [`chansend`](#chansend) but in reverse.
//
// - `ep` is the pointer where to store the received data, it may be nil, in which case received data is ignored. A non-nil `ep` must point to the heap or the caller's stack.
// - `block` comes from the select statemenet, should the receiver block or not? If not then the goroutine will not sleep but return if it could not complete.
//     If block == false and no elements are available, returns (false, false).
//     Otherwise, if c is closed, zeros *ep and returns (true, false).
//     Otherwise, fills in *ep with an element and returns (true, true).
func chanrecv(t *chantype, c *hchan, ep unsafe.Pointer, block bool) (selected, received bool) {
	if debugChan {
		print("chanrecv: chan=", c, "\n")
	}

	//<a name="axiom3"/>
	// **AXIOM #3** - A receive from a nil channel blocks forever
	if c == nil {
		if !block {
			return
		}
		gopark(nil, nil, "chan receive (nil chan)", traceEvGoStop, 2)
		throw("unreachable")
	}

	//*Original Comments:*
	//
	// *Fast path: check for failed non-blocking operation without acquiring the lock.*
	//
	// *After observing that the channel is not ready for receiving, we observe that the
	// channel is not closed. Each of these observations is a single word-sized read
	// (first c.sendq.first or c.qcount, and second c.closed).
	// Because a channel cannot be reopened, the later observation of the channel
	// being not closed implies that it was also not closed at the moment of the
	// first observation. We behave as if we observed the channel at that moment
	// and report that the receive cannot proceed.*
	//
	// *The order of operations is important here: reversing the operations can lead to
	// incorrect behavior when racing with a close.*
	//
	// Fast lock-free check whether there is a chance to receive.
	//
	// If not blocking and the channel is not closed and one of the following conditions are met bail out and return false, we can’t receive at the moment.
	//
	//  1. buffer has zero size and there is no sender (unbuffered channel)
	//  2. buffer has size and is empty (empty buffered channel)
	if !block && (c.dataqsiz == 0 && c.sendq.first == nil ||
		c.dataqsiz > 0 && atomic.Loaduint(&c.qcount) == 0) &&
		atomic.Load(&c.closed) == 0 {
		return
	}

	// Acquire the lock so we are thread safe from now on.
	var t0 int64
	if blockprofilerate > 0 {
		t0 = cputicks()
	}

	lock(&c.lock)

	//<a name="axiom4"/>
	//**AXIOM #4** - A receive from a closed channel returns the zero value immediately
	if c.closed != 0 && c.qcount == 0 {
		if raceenabled {
			raceacquire(unsafe.Pointer(c))
		}
		unlock(&c.lock)
		if ep != nil {
			// Zero the destination memory and return.
			memclr(ep, uintptr(c.elemsize))
		}
		return true, false
	}

	if sg := c.sendq.dequeue(); sg != nil {
		// Found a waiting sender. If buffer is size 0, receive value
		// directly from sender. Otherwise, receive from head of queue
		// and add the sender's value to the tail of the queue (both map to
		// the same buffer slot because the queue is full).[`recv`](#recv)
		recv(c, sg, ep, func() { unlock(&c.lock) })
		return true, true
	}

	// Receive directly from the ring buffer
	if c.qcount > 0 {
		qp := chanbuf(c, c.recvx)
		if raceenabled {
			raceacquire(qp)
			racerelease(qp)
		}
		if ep != nil {
			typedmemmove(c.elemtype, ep, qp)
		}
		// Zero the memory in the ring buffer on the recvx slot.
		memclr(qp, uintptr(c.elemsize))
		// Because it is a ring buffer we wrap the `recvx` if it is
		// pointing beyond the buffer.
		c.recvx++
		if c.recvx == c.dataqsiz {
			c.recvx = 0
		}
		c.qcount--
		// Unlock and return true as the value was received.
		unlock(&c.lock)
		return true, true
	}

	// We are at the point where we couldn't receive any value the
	// ring buffer is empty and there is no sender, so if it should
	// not block bail out now.
	if !block {
		unlock(&c.lock)
		return false, false
	}

	// No sender available: block on this channel.Create new `sudog`
	// struct for current receiver and channel and enqueue it to the
	// wait list.
	gp := getg()
	mysg := acquireSudog()
	mysg.releasetime = 0
	if t0 != 0 {
		mysg.releasetime = -1
	}
	mysg.elem = ep
	mysg.waitlink = nil
	gp.waiting = mysg
	mysg.g = gp
	mysg.selectdone = nil
	mysg.c = c
	gp.param = nil
	c.recvq.enqueue(mysg)
	// Here the goroutine goes to sleep and scheduler will wake it
	// once there is a sender on this channel. The lock is released
	// inside the function.
	goparkunlock(&c.lock, "chan receive", traceEvGoBlockRecv, 3)

	//<a name="rcv_wakeup" />
	// This comes after the blocking, the goroutine is awaken.
	if mysg != gp.waiting {
		throw("G waiting list is corrupted")
	}
	gp.waiting = nil
	if mysg.releasetime > 0 {
		blockevent(mysg.releasetime-t0, 2)
	}
	closed := gp.param == nil
	gp.param = nil
	mysg.c = nil
	releaseSudog(mysg)
	return true, !closed
}

// `recv` processes a receive operation on a full channel c.
// There are 2 steps:
//
// 1. The value sent by the sender `sg` is put into the channel
//    and the sender is woken up to go on its merry way.
// 2. The value received by the receiver (the current G) is
//    written to `ep`.
//
// For synchronous channels, both values are the same.
// For asynchronous channels, the receiver gets its data from
// the channel buffer and the sender's data is put in the
// channel buffer.
//
// Channel c must be full and locked. recv unlocks c with `unlockf`.
// `sg` must already be dequeued from c.
// A non-nil `ep` must point to the heap or the caller's stack.
func recv(c *hchan, sg *sudog, ep unsafe.Pointer, unlockf func()) {
	// There is no buffer, unbuffered channel.
	if c.dataqsiz == 0 {
		if raceenabled {
			racesync(c, sg)
		}
		if ep != nil {
			// Copy data from sender.
			// `ep` points to own stack or heap, so nothing
			// special (ala [`sendDirect`](#sendDirect)) is needed
			// here.
			typedmemmove(c.elemtype, ep, sg.elem)
		}
	} else {
		// The buffer is full. Take the item at the
		// head of the queue. Make the sender enqueue
		// its item at the tail of the queue. Since the
		// queue is full, those are both at the same slot.
		qp := chanbuf(c, c.recvx)
		if raceenabled {
			raceacquire(qp)
			racerelease(qp)
			raceacquireg(sg.g, qp)
			racereleaseg(sg.g, qp)
		}
		// Copy data from queue to receiver
		if ep != nil {
			typedmemmove(c.elemtype, ep, qp)
		}
		// Copy data from sender to queue
		typedmemmove(c.elemtype, qp, sg.elem)
		c.recvx++
		if c.recvx == c.dataqsiz {
			c.recvx = 0
		}
		c.sendx = c.recvx // c.sendx = (c.sendx+1) % c.dataqsiz
	}
	sg.elem = nil
	gp := sg.g
	unlockf()
	// Set the sender `g.param` so it is non-nil to signal that the
	// channel was not closed.
	gp.param = unsafe.Pointer(sg)
	if sg.releasetime != 0 {
		sg.releasetime = cputicks()
	}
	// Wake the sending goroutine
	goready(gp, 4)
}

// ## Select

// `selectnbsend` is used for compiler implementation of
// ```go
//	select {
//	case c <- v:
//		... foo
//	default:
//		... bar
//	}
// ```
// as
//```go
//	if selectnbsend(c, v) {
//		... foo
//	} else {
//		... bar
//	}
//```
func selectnbsend(t *chantype, c *hchan, elem unsafe.Pointer) (selected bool) {
	return chansend(t, c, elem, false, getcallerpc(unsafe.Pointer(&t)))
}

// `selectnbrecv` is used for compiler implementation of
//```go
//	select {
//	case v = <-c:
//		... foo
//	default:
//		... bar
//	}
//```
// as
//```go
//	if selectnbrecv(&v, c) {
//		... foo
//	} else {
//		... bar
//	}
//```
func selectnbrecv(t *chantype, elem unsafe.Pointer, c *hchan) (selected bool) {
	selected, _ = chanrecv(t, c, elem, false)
	return
}

// `selectnbrecv2` is used for compiler implementation of
//```go
//	select {
//	case v, ok = <-c:
//		... foo
//	default:
//		... bar
//	}
//```
// as
//```golang
//	if c != nil && selectnbrecv2(&v, &ok, c) {
//		... foo
//	} else {
//		... bar
//	}
//```
func selectnbrecv2(t *chantype, elem unsafe.Pointer, received *bool, c *hchan) (selected bool) {
	selected, *received = chanrecv(t, c, elem, false)
	return
}

//`reflect.chansend`
func reflect_chansend(t *chantype, c *hchan, elem unsafe.Pointer, nb bool) (selected bool) {
	return chansend(t, c, elem, !nb, getcallerpc(unsafe.Pointer(&t)))
}

//`reflect.chanrecv`
func reflect_chanrecv(t *chantype, c *hchan, nb bool, elem unsafe.Pointer) (selected bool, received bool) {
	return chanrecv(t, c, elem, !nb)
}

//`reflect.chanlen`
func reflect_chanlen(c *hchan) int {
	if c == nil {
		return 0
	}
	return int(c.qcount)
}

//`reflect.chancap`
func reflect_chancap(c *hchan) int {
	if c == nil {
		return 0
	}
	return int(c.dataqsiz)
}

//`reflect.chanclose`
func reflect_chanclose(c *hchan) {
	closechan(c)
}

// Enqueue the `sudog` to the waiters linked list
func (q *waitq) enqueue(sgp *sudog) {
	sgp.next = nil
	x := q.last
	if x == nil {
		sgp.prev = nil
		q.first = sgp
		q.last = sgp
		return
	}
	sgp.prev = x
	x.next = sgp
	q.last = sgp
}

// Dequeue the `sudog` from the waiters linked list
func (q *waitq) dequeue() *sudog {
	for {
		sgp := q.first
		if sgp == nil {
			return nil
		}
		y := sgp.next
		if y == nil {
			q.first = nil
			q.last = nil
		} else {
			y.prev = nil
			q.first = y
			sgp.next = nil // mark as removed (see dequeueSudog)
		}

		if sgp.selectdone != nil {
			if *sgp.selectdone != 0 || !atomic.Cas(sgp.selectdone, 0, 1) {
				continue
			}
		}

		return sgp
	}
}

func racesync(c *hchan, sg *sudog) {
	racerelease(chanbuf(c, 0))
	raceacquireg(sg.g, chanbuf(c, 0))
	racereleaseg(sg.g, chanbuf(c, 0))
	raceacquire(chanbuf(c, 0))
}
