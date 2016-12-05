package zplprinter

import (
	"runtime"

	"github.com/giddyinc/gousb/usb"
)

type UsbConfig struct {
	Vendor   usb.ID
	Product  usb.ID
	Config   uint8
	Iface    uint8
	Setup    uint8
	Endpoint uint8
}

type UsbZplPrinter struct {
	*usb.Device
	Config UsbConfig
}

func (printer *UsbZplPrinter) Write(buf []byte) (int, error) {
	endpoint, err := printer.OpenEndpoint(
		printer.Config.Config,
		printer.Config.Iface,
		printer.Config.Setup,
		printer.Config.Endpoint|uint8(usb.ENDPOINT_DIR_OUT),
	)
	if err != nil {
		return 0, err
	}

	l, err := endpoint.Write(buf)
	return l, err
}

func GetPrinters(ctx *usb.Context, config UsbConfig) ([]*UsbZplPrinter, error) {
	var printers []*UsbZplPrinter
	devices, err := ctx.ListDevices(func(desc *usb.Descriptor) bool {
		var selected = desc.Vendor == config.Vendor
		if config.Product != usb.ID(0) {
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
						if end.Direction() == usb.ENDPOINT_DIR_OUT {
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
