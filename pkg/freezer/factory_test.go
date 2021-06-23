package freezer

import (
	"context"
	"testing"
)

func TestDocker_Freeze(t *testing.T) {
	f, err := GetFreezer(RuntimeTypeDocker)
	if err != nil {
		t.Errorf("Get freezer error: %s", err.Error())
	}
	err = f.Freeze(context.TODO(), "1e500cda-a036-4aa8-bc3b-88daeeff10e3", "eventing-controller")
	if err != nil {
		t.Errorf("error when freeze,err: %s", err.Error())
	}
}

func TestDocker_Thaw(t *testing.T) {
	f, err := GetThawer(RuntimeTypeDocker)
	if err != nil {
		t.Errorf("Get freezer error: %s", err.Error())
	}
	err = f.Thaw(context.TODO(), "1e500cda-a036-4aa8-bc3b-88daeeff10e3", "eventing-controller")
	if err != nil {
		t.Errorf("error when freeze,err: %s", err.Error())
	}
}
