package main

import (
	"github.com/google/uuid"
)

// ----------------------------------------------------------------------------
// h e l p e r   f u n c t i o n s
// ----------------------------------------------------------------------------

func IsValidUUID(u string) bool {
    _, err := uuid.Parse(u)
    return err == nil
}
