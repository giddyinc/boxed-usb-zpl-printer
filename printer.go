package usbzplprinter

import (
	"runtime"

	"github.com/google/gousb"
)

type UsbConfig struct {
	Vendor   gousb.ID
	Product  gousb.ID
	Config   uint8
	Iface    uint8
	Setup    uint8
	Endpoint uint8
}

type UsbZplPrinter struct {
	*gousb.Device
	Config UsbConfig
}

func (printer *UsbZplPrinter) Write(buf []byte) (int, error) {
	endpoint, err := printer.OpenEndpoint(
		printer.Config.Config,
		printer.Config.Iface,
		printer.Config.Setup,
		printer.Config.Endpoint|uint8(gusb.ENDPOINT_DIR_OUT),
	)
	if err != nil {
		return 0, err
	}

	l, err := endpoint.Write(buf)
	return l, err
}

func GetPrinters(ctx *gusb.Context, config UsbConfig) ([]*UsbZplPrinter, error) {
	var printers []*UsbZplPrinter
	devices, err := ctx.ListDevices(func(desc *gusb.Descriptor) bool {
		var selected = desc.Vendor == config.Vendor
		if config.Product != gusb.ID(0) {
			selected = selected && desc.Product == config.Product
		}
		return selected
	})

	if err != nil {
		return printers, err
	}

	if len(devices) == 0 {
		return printers, ErrorDeviceNotFound
	}

getDevice:
	for _, dev := range devices {
		if runtime.GOOS == "linux" {
			dev.DetachKernelDriver(0)
		}

		// get devices with IN direction on endpoint
		for _, cfg := range dev.Descriptor.Configs {
			for _, alt := range cfg.Interfaces {
				for _, iface := range alt.Setups {
					for _, end := range iface.Endpoints {
						if end.Direction() == gusb.ENDPOINT_DIR_OUT {
							config.Config = cfg.Config
							config.Iface = alt.Number
							config.Setup = iface.Number
							config.Endpoint = uint8(end.Number())
							printer := &UsbZplPrinter{
								dev,
								config,
							}
							// don't timeout reading
							printer.ReadTimeout = 0
							printers = append(printers, printer)
							continue getDevice
						}
					}
				}
			}
		}
	}

	return printers, nil
}
