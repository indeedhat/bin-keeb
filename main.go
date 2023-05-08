package main

import (
	"machine"
	"machine/usb/hid/keyboard"
	"time"
)

const (
	// When all keys are released this threshold will be used to determine what bits are set
	// any key that has been pressed within the threshold will have its bit counted
	Timeout = 50 * time.Millisecond

	// print out the binary representation of the
	Debug = false
)

var buttons = []machine.Pin{
	machine.GPIO22,
	machine.GPIO28,
	machine.GPIO27,
	machine.GPIO26,
	machine.GPIO9,
	machine.GPIO8,
	machine.GPIO7,
	machine.GPIO6,
}

func main() {
	initButtons()

	var (
		keeb = keyboard.Port()
		buf  = make([]time.Time, 8)
	)

	for {
		time.Sleep(10 * time.Millisecond)

		var keyCount uint8

		for i := 0; i < len(buttons); i++ {
			keyCount += logPress(&buf[i], buttons[i].Get())
		}

		// if no keys are being pressed then there is no need to do anything else this iteration
		if keyCount != 0 {
			continue
		}

		if Debug {
			bits, write := buildStateDebug(buf)
			if !write {
				continue
			}

			keeb.Write(bits)
		} else {
			state := buildState(buf)
			if state == 0 {
				continue
			}

			keeb.WriteByte(state)
		}

		buf = make([]time.Time, 8)
	}
}

// logPress logs the time of a key that is currently being pressed
func logPress(t *time.Time, mark bool) uint8 {
	if !mark {
		return 0
	}

	*t = time.Now()
	return 1

}

// initButtons sets up the pin state for each of the keyboard buttons
func initButtons() {
	for _, btn := range buttons {
		btn.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	}
}

// buildState builds the final key state based on the keypresses in the expected threshold
func buildState(buf []time.Time) (state uint8) {
	threshold := time.Now().Add(-Timeout)

	for i, pressTime := range buf {
		if pressTime.Before(threshold) {
			continue
		}

		state = state | 1<<(7-i)
	}

	return
}

// buildStateDebug builds up a string to send that will display the binary representation of the
// key state on keyup
func buildStateDebug(buf []time.Time) (state []byte, has bool) {
	threshold := time.Now().Add(-Timeout)

	state = make([]byte, 8)

	for i, pressTime := range buf {
		if pressTime.Before(threshold) {
			state[i] = '0'
		} else {
			state[i] = '1'
			has = true
		}
	}

	return
}
