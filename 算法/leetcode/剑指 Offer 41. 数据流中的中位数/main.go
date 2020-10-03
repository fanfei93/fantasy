package  main


type MedianFinder struct {
	minHeap []int
	maxHeap []int
	count int
}


/** initialize your data structure here. */
func Constructor() MedianFinder {
	m := MedianFinder{
		minHeap:[]int{},
		maxHeap:[]int{},
		count:0,
	}
	return m
}


func (this *MedianFinder) AddNum(num int)  {
	this.count++
	if this.count == 1 {
		this.minHeap = append(this.minHeap, num)
	} else {
		if len(this.minHeap) == len(this.maxHeap) {
			if this.minHeap[0] > num {
				this.maxHeap = append(this.maxHeap, this.minHeap[0])
				rebuild(this.maxHeap)
				//this.minHeap[0] = num
				//heapify(this.minHeap, 0, false)
			} else {
				this.minHeap[0] = num
				heapify(this.minHeap, 0, false)

				this.maxHeap = append(this.maxHeap, num)
				rebuild(this.maxHeap)
			}
		} else {
			if this.minHeap[0] > num {
				this.maxHeap = append(this.maxHeap, num)
				rebuild(this.minHeap)
			} else {
				this.maxHeap = append(this.maxHeap,this.minHeap[0])
				rebuild(this.maxHeap)
				this.minHeap[0]  = num
				heapify(this.minHeap,0, false)
			}
		}
	}
}


func (this *MedianFinder) FindMedian() float64 {
	var res float64

	return res
}

func heapify(nums []int, start int, isMax bool) {
	left := 2 * start + 1
	right := 2 * start + 2

	min := start
	if left < len(nums) {
		if isMax  {

		}
	}
}

func rebuild(nums []int) {

}

/**
 * Your MedianFinder object will be instantiated and called as such:
 * obj := Constructor();
 * obj.AddNum(num);
 * param_2 := obj.FindMedian();
 */