package reporter

import (
	"github.com/ncaunt/meldec/internal/pkg/decoder"
)

type Reporter interface {
	Publish(decoder.Stat)
}
