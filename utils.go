package cache

import "time"

func MathClampF32(val, min, max float32) float32 {
    if val < min {
        return min
    } else if val > max {
        return max
    } else {
        return val
    }
}

var GetTime = time.Now
