package main

type ParkingSystem struct {
	bigCap int
	bigExist int
	mediumCap int
	mediumExist  int
	smallCap int
	smallExist int
}


func Constructor(big int, medium int, small int) ParkingSystem {
	return ParkingSystem{
		bigCap:      big,
		bigExist:    0,
		mediumCap:   medium,
		mediumExist: 0,
		smallCap:    small,
		smallExist:  0,
	}
}


func (this *ParkingSystem) AddCar(carType int) bool {
	cap := this.getCap(carType)
	exist := this.getExist(carType)
	if cap == exist {
		return false
	} else {
		if carType == 1 {
			this.bigExist++
		} else if carType == 2 {
			this.mediumExist++
		} else {
			this.smallExist++
		}
		return true
	}
}

func (this *ParkingSystem) getCap(carType int) int {
	if carType == 1 {
		return this.bigCap
	} else if carType == 2 {
		return this.mediumCap
	} else {
		return this.smallCap
	}
}

func (this *ParkingSystem) getExist(carType int) int {
	if carType == 1 {
		return this.bigExist
	} else if carType == 2 {
		return this.mediumExist
	} else {
		return this.smallExist
	}
}



/**
 * Your ParkingSystem object will be instantiated and called as such:
 * obj := Constructor(big, medium, small);
 * param_1 := obj.AddCar(carType);
 */
