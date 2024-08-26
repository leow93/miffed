// TODO: delete this file when done
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/leow93/miffed-api/internal/liftv3"
)

func callLift(svc *liftv3.LiftService, id liftv3.LiftId, floor int) {
	ctx := context.TODO()
	err := svc.CallLift(ctx, id, floor)
	if err != nil {
		panic(err)
	}
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	svc := liftv3.NewLiftService(ctx)

	liftCount := 1

	var lifts []liftv3.Lift

	for range liftCount {
		lift, err := svc.AddLift(liftv3.LiftConfig{Floor: 0})
		if err != nil {
			panic(err)
		}
		lifts = append(lifts, lift)
	}

	fmt.Println("lifts", lifts)

	for _, l := range lifts {
		callLift(svc, l.Id, 5)
		callLift(svc, l.Id, 5)
		callLift(svc, l.Id, 5)
		callLift(svc, l.Id, 5)
		callLift(svc, l.Id, 5)
		callLift(svc, l.Id, 5)
		callLift(svc, l.Id, 6)
		callLift(svc, l.Id, 10)
		callLift(svc, l.Id, 10000)
	}

	time.Sleep(time.Second)

	l, err := svc.AddLift(liftv3.LiftConfig{Floor: 0})
	if err != nil {
		panic(err)
	}
	callLift(svc, l.Id, 1)
	callLift(svc, l.Id, 3)
	callLift(svc, l.Id, 17)
	cancel()
	callLift(svc, l.Id, 99)
	time.Sleep(time.Second * 2)
}
