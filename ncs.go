package ncs

// #cgo LDFLAGS: -lmvnc
/*
#include <ncs.h>
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// StatusCode is the NCSDK API status code as returned by most API calls.
// It usually reports the status of the neural compute stick.
type StatusCode int

const (
	// StatusOK means the API function call worked as expected
	StatusOK StatusCode = 0

	// StatusBusy means device is busy, retry later.
	StatusBusy = -1

	// StatusError means an unexpected error was encountered during the API function call.
	StatusError = -2

	// StatusOutOfMemory means the host is out of memory.
	StatusOutOfMemory = -3

	// StatusDeviceNotFound means no device has been found at the given index or name.
	StatusDeviceNotFound = -4

	// StatusInvalidParameters means at least one of the given parameters is wrong.
	StatusInvalidParameters = -5

	// StatusTimeout means there was a timeout in the communication with the device.
	StatusTimeout = -6

	// StatusCmdNotFound means the file to boot the device was not found.
	StatusCmdNotFound = -7

	// StatusNotAllocated means the graph or fifo has not been allocated.
	StatusNotAllocated = -8

	// StatusUnauthorized means an unauthorized operation has been attempted.
	StatusUnauthorized = -9

	// StatusUnsupportedGraphFile means the graph file version is not supported.
	StatusUnsupportedGraphFile = -10

	// StatusUnsupportedConfigFile is reserved for future use.
	StatusUnsupportedConfigFile = -11

	// StatusUnsupportedFeature means the operation used a feature unsupported by this firmware version.
	StatusUnsupportedFeature = -12

	// StatusMyriadError when an error has been reported by device, use MVNC_DEBUG_INFO.
	StatusMyriadError = -13

	// StatusInvalidDataLength means an invalid data length has been passed when getting or setting an option
	StatusInvalidDataLength = -14

	// StatusInvalidHandle means an invalid handle has been passed to a function
	StatusInvalidHandle = -15
)

// String method to satisfy fmt.Stringer interface
func (nc StatusCode) String() string {
	switch nc {
	case StatusOK:
		return "OK"
	case StatusBusy:
		return "Device busy"
	case StatusError:
		return "Unexpected error"
	case StatusOutOfMemory:
		return "Host out of memory"
	case StatusDeviceNotFound:
		return "Device not found"
	case StatusInvalidParameters:
		return "Invalid parameters"
	case StatusTimeout:
		return "Device timeout"
	case StatusCmdNotFound:
		return "Device bootloader not found"
	case StatusNotAllocated:
		return "Unallocated resource"
	case StatusUnauthorized:
		return "Unauthorized operation"
	case StatusUnsupportedGraphFile:
		return "Unsupported graph file"
	case StatusUnsupportedConfigFile:
		return "Unsupported configuration"
	case StatusUnsupportedFeature:
		return "Unsupported feature"
	case StatusMyriadError:
		return "Movidius VPU failure"
	case StatusInvalidDataLength:
		return "Invalid data length when setting options"
	case StatusInvalidHandle:
		return "Invalid handle"
	default:
		return "Unknown status"
	}
}

// Device is Neural Compute Stick (NCS) device
type Device struct {
	handle unsafe.Pointer
}

// NewDevice creates new NCS device handle and returns it
//
// For more information:
// https://movidius.github.io/ncsdk/ncapi/ncapi2/c_api/ncDeviceCreate.html
func NewDevice(index int) (*Device, error) {
	var handle unsafe.Pointer

	c := C.ncs_DeviceCreate(C.int(index), &handle)

	if StatusCode(c) != StatusOK {
		return nil, fmt.Errorf("Failed to create new device: %s", StatusCode(c))
	}

	return &Device{handle: handle}, nil
}

// Open initializes NCS device and opens device communication channel
// It returns error if it fails to open the device.
//
// For more information:
// https://movidius.github.io/ncsdk/ncapi/ncapi2/c_api/ncDeviceOpen.html
func (d *Device) Open() error {
	c := C.ncs_DeviceOpen(d.handle)

	if StatusCode(c) != StatusOK {
		return fmt.Errorf("Failed to open device: %s", StatusCode(c))
	}

	return nil
}

// Close closes the communication channel with NCS device.
// It returns error if it fails to close the communication channel.
//
// For more information:
// https://movidius.github.io/ncsdk/ncapi/ncapi2/c_api/ncDeviceClose.html
func (d *Device) Close() error {
	c := C.ncs_DeviceClose(d.handle)

	if StatusCode(c) != StatusOK {
		return fmt.Errorf("Failed to close device: %s", StatusCode(c))
	}

	return nil
}

// Destroy destroys NCS device handle and frees associated resources.
// This function must be called for every device that was initialized with NewDevice().
//
// For more information:
// https://movidius.github.io/ncsdk/ncapi/ncapi2/c_api/ncDeviceDestroy.html
func (d *Device) Destroy() error {
	c := C.ncs_DeviceDestroy(&d.handle)

	if StatusCode(c) != StatusOK {
		return fmt.Errorf("Failed to destroy device: %s", StatusCode(c))
	}

	return nil
}

// Graph is NCSDK neural network graph
type Graph struct {
	name    string
	handle  unsafe.Pointer
	device  *Device
	inFifo  *Fifo
	outFifo *Fifo
}

// NewGraph creates new Graph with given name and returns it
// It returns error if it fails to create new graph
//
// For more information:
// https://movidius.github.io/ncsdk/ncapi/ncapi2/c_api/ncGraphCreate.html
func NewGraph(name string) (*Graph, error) {
	var handle unsafe.Pointer

	_name := C.CString(name)
	defer C.free(unsafe.Pointer(_name))

	c := C.ncs_GraphCreate(_name, &handle)

	if StatusCode(c) != StatusOK {
		return nil, fmt.Errorf("Failed to create new graph: %s", StatusCode(c))
	}

	return &Graph{name: name, handle: handle}, nil
}

// Allocate allocates a graph on NCS device. This function sends graphData to NCS device. It does not allocate input or output FIFO queues. You have to either allocate them separately or use either AllocateWithFifos() or AllocateWithFifosEx() functions whcih conveniently create and allocate the FIFO queues.
// It returns error if it fails to allocate the graph on the device
//
// For more information:
// https://movidius.github.io/ncsdk/ncapi/ncapi2/c_api/ncGraphAllocate.html
func (g *Graph) Allocate(d *Device, graphData []byte) error {
	c := C.ncs_GraphAllocate(d.handle, g.handle, unsafe.Pointer(&graphData[0]), C.uint(len(graphData)))

	if StatusCode(c) != StatusOK {
		return fmt.Errorf("Failed to allocate new graph: %s", StatusCode(c))
	}

	g.device = d

	return nil
}

// AllocateWithFifosDefault allocates a graph and creates and allocates FIFO queues with default parameters for inference. Both FIFOs have FifoDataType set to FifoFP32. Inbound FIFO queue is initialized with FifoHostWO type and outbound FIFO queue with FifoHostRO type. It returns Queue or error if it fails to allocate the graph.
//
// For more information:
// https://movidius.github.io/ncsdk/ncapi/ncapi2/c_api/ncGraphAllocateWithFifos.html
func (g *Graph) AllocateWithFifosDefault(d *Device, graphData []byte) (*Queue, error) {
	return g.AllocateWithFifosOpts(d, graphData, &FifoOpts{FifoHostWO, FifoFP32, 2}, &FifoOpts{FifoHostRO, FifoFP32, 2})
}

// AllocateWithFifosOpts allocates a graph and creates and allocates FIFO queues for inference. This function is similar to AllocateWithFifosDefault, but rather than initializing FIFOs with default values it accepts parameters that allow to specify FIFO queue parameters
//
// For more information:
// https://movidius.github.io/ncsdk/ncapi/ncapi2/c_api/ncGraphAllocateWithFifosEx.html
func (g *Graph) AllocateWithFifosOpts(d *Device, graphData []byte, inOpts *FifoOpts, outOpts *FifoOpts) (*Queue, error) {
	var inHandle, outHandle unsafe.Pointer

	c := C.ncs_GraphAllocateWithFifosEx(d.handle,
		g.handle, unsafe.Pointer(&graphData[0]), C.uint(len(graphData)),
		&inHandle, C.ncFifoType(inOpts.Type), C.int(inOpts.NumElem), C.ncFifoDataType(inOpts.DataType),
		&outHandle, C.ncFifoType(outOpts.Type), C.int(outOpts.NumElem), C.ncFifoDataType(outOpts.DataType))

	if StatusCode(c) != StatusOK {
		return nil, fmt.Errorf("Failed to allocate graph with FIFOs: %s", StatusCode(c))
	}

	inFifo := &Fifo{handle: inHandle, device: d}
	outFifo := &Fifo{handle: outHandle, device: d}

	g.device = d
	g.inFifo = inFifo
	g.outFifo = outFifo

	return &Queue{
		In:  g.inFifo,
		Out: g.outFifo,
	}, nil
}

// Destroy destroys NCS graph handle and frees associated resources.
// This function must be called for every graph that was initialized with NewGraph().
//
// For more information:
// https://movidius.github.io/ncsdk/ncapi/ncapi2/c_api/ncGraphDestroy.html
func (g *Graph) Destroy() error {
	c := C.ncs_GraphDestroy(&g.handle)

	if StatusCode(c) != StatusOK {
		return fmt.Errorf("Failed to destroy graph: %s", StatusCode(c))
	}

	return nil
}

// FifoType defines FIFO access types
//
// For more information:
// https://movidius.github.io/ncsdk/ncapi/ncapi2/c_api/ncFifoType_t.html
type FifoType int

const (
	// FifoHostRO allows Read Only API access and Read-Write Graph access
	FifoHostRO FifoType = 0
	// FifoHostWO allows Write Only API acess and Read Only Graph access
	FifoHostWO = 1
)

// Queue is a queue used for NCS inference
type Queue struct {
	// In is an inbound queue
	In *Fifo
	// Out is an outbound queue
	Out *Fifo
}

// FifoDataType defines possible data types for FIFOs
//
// For more information:
// https://movidius.github.io/ncsdk/ncapi/ncapi2/c_api/ncFifoDataType_t.html
type FifoDataType int

const (
	// FifoFP16 data is in half precision (16 bit) floating point format (FP16).
	FifoFP16 FifoDataType = 0
	// FifoFP32 data is in full precision (32 bit) floating point format (FP32)
	FifoFP32 = 1
)

// FifoOpts specifies FIFO configuration options
// FIXME: This will most likely change soon
type FifoOpts struct {
	// Type is FIFO type
	Type FifoType
	// DataType is FIFO data type
	DataType FifoDataType
	// NumElem is a max number of elements that the FIFO will be able to contain
	NumElem int
}

// Fifo is NCSDK FIFO queue
type Fifo struct {
	name   string
	handle unsafe.Pointer
	device *Device
}

// NewFifo creates new FIFO queue with given name and returns it
// It returns error if it fails to create new queue
//
// For more information:
// https://movidius.github.io/ncsdk/ncapi/ncapi2/c_api/ncFifoCreate.html
func NewFifo(name string, t FifoType) (*Fifo, error) {
	var handle unsafe.Pointer

	_name := C.CString(name)
	defer C.free(unsafe.Pointer(_name))

	c := C.ncs_FifoCreate(_name, C.ncFifoType(t), &handle)

	if StatusCode(c) != StatusOK {
		return nil, fmt.Errorf("Failed to create new FIFO: %s", StatusCode(c))
	}

	return &Fifo{name: name, handle: handle}, nil
}

// Destroy destroys NCS FIFO handle and frees associated resources.
// This function must be called for every FIFO handle that was initialized with NewFifo()
//
// For more information:
// https://movidius.github.io/ncsdk/ncapi/ncapi2/c_api/ncFifoDestroy.html
func (f *Fifo) Destroy() error {
	c := C.ncs_FifoDestroy(&f.handle)

	if StatusCode(c) != StatusOK {
		return fmt.Errorf("Failed to destroy FIFO: %s", StatusCode(c))
	}

	return nil
}
