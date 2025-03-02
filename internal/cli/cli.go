// cli/cli.go

package cli

import (
	"OZON/internal/repository/memory"
	postgres "OZON/internal/repository/postrges"
	"OZON/pkg/storage"
	"flag"
	"log"
)

// Flag представляет флаги для выбора типа хранения
type Flag struct {
	storageType string
}

func (f *Flag) ParseFlag() {
	flag.StringVar(&f.storageType, "storage", "postgres", "Storage type: postgres or inmemory (default: postgres)")
	flag.Parse()
}

func ProcessFlag(db storage.DB) interface{} {
	f := &Flag{}
	f.ParseFlag()

	switch f.storageType {
	case "inmemory":
		return memory.NewInMemoryRepository()
	case "postgres":
		return postgres.NewPostgresRepository(db)
	default:
		log.Fatalf("unknown storage type: %s. Use 'postgres' or 'inmemory'", f.storageType)
		return nil
	}
}
