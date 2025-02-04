# go-usbtemp

This library is designed to be used with DS18B20 1-wire sensors such as those available on https://usbtemp.com

Code for this library is heavily influences by the examples available here: https://github.com/usbtemp

See the example in the cmd folder for a more thorough example of its usage, but basically this should work:

```
package main

import (
  "fmt"
  "github.com/jazzboME/go-usbtemp"
}

func main() {
  var probe = usbtemp.USBtemp{}

  probe.Open("/dev/ttyUSB0")
  defer probe.Close()

  temp, _ := probe.Temperature(true)  // returns Fahrenheit; pass false for Celsius.
  fmt.Printf("Current temperature is: %.2f\n", temp)
}
```

In addition to `Open()`, `Close()` and `Temperature()`, `Rom()` is also available.

The `USBtemp` struct also exposes the USB ID and Serial Number.
