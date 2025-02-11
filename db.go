package main

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/dgraph-io/badger/v3"
)

func InitDB(database string) (*badger.DB, error) {
	opts := badger.DefaultOptions(database)
	opts.Logger = nil // Deshabilita el logging

	opts.SyncWrites = true
	opts.NumMemtables = 3
	opts.NumLevelZeroTables = 5
	opts.NumLevelZeroTablesStall = 10

	// Gestión de memoria
	opts.CompactL0OnClose = true

	// Configuraciones de memoria más modernas
	opts.MemTableSize = 64 << 20      // 64MB tamaño de memtable
	opts.ValueLogFileSize = 256 << 20 // 256MB tamaño de archivo de registro de valores

	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("error abriendo base de datos %s: %v", database, err)
	}
	return db, nil
}

// Crear (Insertar)
func InsertKV(db *badger.DB, key string, value []byte) error {
	return db.Update(func(txn *badger.Txn) error {
		// Primero verificamos si la clave existe
		_, err := txn.Get([]byte(key))
		if err == nil {
			// La clave existe
			return fmt.Errorf("la clave '%s' ya existe", key)
		} else if err != badger.ErrKeyNotFound {
			// Error diferente al de "clave no encontrada"
			return err
		}

		// La clave no existe, procedemos a insertarla
		return txn.Set([]byte(key), value)
	})
}

func GetAllKeys(db *badger.DB) (map[string]any, error) {
	var results = make(map[string]any)
	err := db.View(func(txn *badger.Txn) error {
		// Crear un iterador para recorrer todas las claves en la base de datos
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.Key()

			// Obtener el valor de la clave
			err := item.Value(func(val []byte) error {
				// Agregar el par clave-valor al slice
				var datos map[string]any
				json.Unmarshal(val, &datos)
				results[string(key)] = datos
				return nil
			})

			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return results, nil
}

// Leer
func SelectKV(db *badger.DB, key string) ([]byte, error) {
	var value []byte
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		return item.Value(func(v []byte) error {
			value = append(value, v...)
			return nil
		})
	})
	return value, err
}

// Actualizar
func UpdateKV(db *badger.DB, key string, newValue []byte) error {
	return db.Update(func(txn *badger.Txn) error {
		// Verifica si la clave existe antes de actualizar
		_, err := txn.Get([]byte(key))
		if err != nil {
			return err // Retorna error si la clave no existe
		}
		return txn.Set([]byte(key), newValue)
	})
}

// Eliminar
func DeleteVK(db *badger.DB, key string) error {
	return db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

func Find(mapa map[string]string, expresion string) (map[string]string, error) {
	// Compilar la expresión regular y manejar el posible error
	re, err := regexp.Compile(expresion)
	if err != nil {
		return nil, err // Devolver el error si la expresión es inválida
	}

	// Crear un nuevo mapa para almacenar los resultados
	resultados := make(map[string]string)
	for clave, valor := range mapa {
		if re.MatchString(clave) {
			resultados[clave] = valor
		}
	}
	return resultados, nil
}
