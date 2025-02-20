package telegram

const (
	stateDefault int16 = iota
	stateGarden        = iota << 8

	stateMaskStep    = 0b0000_0000_1111
	stateMaskCommand = 0b0000_1111_0000
	stateMaskGroup   = 0b1111_0000_0000
)

// TODO add a command registration function to wrap the any registered commands with a stateDefault check
