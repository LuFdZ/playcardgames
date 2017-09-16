package date

import "time"

func TimeSubDays(t1, t2 time.Time) int {
	// if t1.Location().String() != t2.Location().String() {
	// 	return -1
	// }
	hours := t1.Sub(t2).Hours()

	if hours <= 0 {
		return -1
	}
	// sub hours less than 24
	if hours < 24 {
		// may same day
		t1y, t1m, t1d := t1.Date()
		t2y, t2m, t2d := t2.Date()
		isSameDay := (t1y == t2y && t1m == t2m && t1d == t2d)
		if isSameDay {
			return 0
		} else {
			return 1
		}
	} else { // equal or more than 24
		if (hours/24)-float64(int(hours/24)) == 0 { // just 24's times
			return int(hours / 24)
		} else { // more than 24 hours
			return int(hours/24) + 1
		}
	}
}

// // just one second
// t1, _ := time.Parse(layout, "2007-01-02 23:59:59")
// t2, _ := time.Parse(layout, "2007-01-03 00:00:00")
// if timeSub(t2, t1) != 1 {
//     panic("one second but different day should return 1")
// }

// // just one day
// t1, _ = time.Parse(layout, "2007-01-02 23:59:59")
// t2, _ = time.Parse(layout, "2007-01-03 23:59:59")
// if timeSub(t2, t1) != 1 {
//     panic("just one day should return 1")
// }

// // more than one day
// t1, _ = time.Parse(layout, "2007-01-02 23:59:59")
// t2, _ = time.Parse(layout, "2007-01-04 00:00:00")
// if timeSub(t2, t1) != 2 {
//     panic("just one day should return 2")
// }
// // just 3 day
// t1, _ = time.Parse(layout, "2007-01-02 00:00:00")
// t2, _ = time.Parse(layout, "2007-01-05 00:00:00")
// if timeSub(t2, t1) != 3 {
//     panic("just 3 day should return 3")
// }

// // different month
// t1, _ = time.Parse(layout, "2007-01-02 00:00:00")
// t2, _ = time.Parse(layout, "2007-02-02 00:00:00")
// if timeSub(t2, t1) != 31 {
//     fmt.Println(timeSub(t2, t1))
//     panic("just one month:31 days should return 31")
// }

// // 29 days in 2mth
// t1, _ = time.Parse(layout, "2000-02-01 00:00:00")
// t2, _ = time.Parse(layout, "2000-03-01 00:00:00")
// if timeSub(t2, t1) != 29 {
//     fmt.Println(timeSub(t2, t1))
//     panic("just one month:29 days should return 29")
// }
