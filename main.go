package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

//var htmlContent embed.FS

type Consulta struct {
	DbType    string `json:"DbType"`
	Dbname    string `json:"Dbname"`
	Apikey    string `json:"Apikey"`
	Querytype string `json:"Querytype"`
	Dbquery   string `json:"Dbquery"`
	Params    []any  `json:"Params"`
}

type BackupConfig struct {
	User      string
	Password  string
	Host      string
	Port      string
	Database  string
	BackupDir string
}

func PrintGreen(text ...string) {
	fondoVerde := color.New(color.FgBlack, color.BgGreen)
	resultado := strings.Join(text, " ")
	fondoVerde.Println(resultado)
}

func getTables(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}
	return tables, nil
}

func getCreateTable(db *sql.DB, table string) (string, error) {
	var tableCreate string
	err := db.QueryRow("SHOW CREATE TABLE "+table).Scan(&table, &tableCreate)
	if err != nil {
		return "", err
	}
	return tableCreate + ";\n", nil
}

func hasRows(db *sql.DB, table string) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func dumpTable(db *sql.DB, table string, file *os.File) error {
	// Obtener estructura de la tabla
	create, err := getCreateTable(db, table)
	if err != nil {
		return err
	}
	file.WriteString("\n-- Estructura de tabla: " + table + "\n")
	file.WriteString("DROP TABLE IF EXISTS `" + table + "`;\n")
	file.WriteString(create)

	// Verificar si la tabla tiene registros
	hasData, err := hasRows(db, table)
	if err != nil {
		return err
	}

	if !hasData {
		return nil // Salir si no hay datos
	}

	// Obtener datos
	rows, err := db.Query("SELECT * FROM " + table)
	if err != nil {
		return err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	// Preparar buffer para los datos
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range columns {
		valuePtrs[i] = &values[i]
	}

	file.WriteString("\n-- Datos de tabla: " + table + "\n")
	file.WriteString("INSERT INTO `" + table + "` VALUES\n")

	var rowCount int
	for rows.Next() {
		err := rows.Scan(valuePtrs...)
		if err != nil {
			return err
		}

		if rowCount > 0 {
			file.WriteString(",\n")
		}

		// Construir valores de inserción
		var rowValues []string
		for _, val := range values {
			if val == nil {
				rowValues = append(rowValues, "NULL")
			} else {
				switch v := val.(type) {
				case []byte:
					rowValues = append(rowValues, fmt.Sprintf("'%s'", strings.ReplaceAll(string(v), "'", "\\'")))
				case string:
					rowValues = append(rowValues, fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "\\'")))
				case time.Time:
					rowValues = append(rowValues, fmt.Sprintf("'%s'", v.Format("2006-01-02 15:04:05")))
				default:
					rowValues = append(rowValues, fmt.Sprintf("%v", v))
				}
			}
		}
		file.WriteString("(" + strings.Join(rowValues, ",") + ")")
		rowCount++
	}
	if rowCount > 0 {
		file.WriteString(";\n")
	}
	return nil
}

func createBackup(config BackupConfig) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("error conectando a la base de datos: %v", err)
	}
	defer db.Close()

	// Verificar conexión
	if err := db.Ping(); err != nil {
		return fmt.Errorf("error verificando conexión: %v", err)
	}

	// Crear archivo de backup
	timestamp := time.Now().Format("2006-01-02")
	filename := filepath.Join(config.BackupDir, fmt.Sprintf("%s_%s.sql", config.Database, timestamp))

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creando archivo: %v", err)
	}
	defer file.Close()

	// Escribir encabezado
	file.WriteString(fmt.Sprintf("-- Backup de %s\n", config.Database))
	file.WriteString(fmt.Sprintf("-- Fecha: %s\n\n", timestamp))
	file.WriteString(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`;\n", config.Database))
	file.WriteString(fmt.Sprintf("USE `%s`;\n\n", config.Database))

	// Obtener lista de tablas
	tables, err := getTables(db)
	if err != nil {
		return fmt.Errorf("error obteniendo tablas: %v", err)
	}

	// Respaldar cada tabla
	for _, table := range tables {
		if err := dumpTable(db, table, file); err != nil {
			return fmt.Errorf("error respaldando tabla %s: %v", table, err)
		}
		fmt.Println("Respaldo de: ", table)
	}

	log.Printf("Backup creado exitosamente: %s", filename)
	PrintGreen("BACKUP CREADO EXITOSAMENTE")
	return nil
}

func respaldo(conf map[string]string) {
	PrintGreen("Iniciando respaldo de base de datos en la carpeta static...")
	config := BackupConfig{
		User:      conf["dbuser"],
		Password:  conf["dbpass"],
		Host:      conf["dbhost"],
		Port:      "3306",
		Database:  conf["dbname"],
		BackupDir: "static",
	}

	// Crear directorio si no existe
	if err := os.MkdirAll(config.BackupDir, 0755); err != nil {
		log.Fatal(err)
	}

	// Ticker para ejecutar cada hora
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	// Primera ejecución inmediata
	if err := createBackup(config); err != nil {
		log.Printf("Error en backup: %v", err)
	}

	// Loop principal
	for range ticker.C {
		if err := createBackup(config); err != nil {
			log.Printf("Error en backup: %v", err)
		}
	}
	PrintGreen("BACKUP CREADO EXITOSAMENTE")
}

func ejeBakup(conf map[string]string) {
	//RespaldoJson()
	respaldo(conf)
	for {
		time.Sleep(5 * time.Hour)
		//RespaldoJson()
		respaldo(conf)
	}
}

func main() {
	CreateConfig()
	//ExtractEmbeddedFiles()
	confs, _ := LoadConfs()
	//CreateTorrc(confs["port"], confs["tor"])
	//go executeTor()
	if confs["dbtype"] == "mysql" {
		go ejeBakup(confs)
	}
	r := GinRouter()

	r.POST("/", func(c *gin.Context) {
		var datos map[string]interface{}
		if err := c.ShouldBindJSON(&datos); err != nil {
			fmt.Println(err)
			c.JSON(http.StatusOK, gin.H{"status": "error", "message": err.Error()})
			return
		}
		fmt.Println(datos)
		if confs["apikey"] != datos["apikey"].(string) {
			fmt.Println("apikey")
			c.JSON(http.StatusOK, gin.H{"status": "error", "message": "invalid apikey"})
			return
		}

		if datos["dbtype"].(string) == "sqlite3" {
			if datos["querytype"].(string) == "select" {
				params := make([]any, 0)
				if datos["params"] != nil {
					params = datos["params"].([]any)
				}
				if len(params) > 0 {
					dat, err := AssocSecure(datos["dbname"].(string), datos["dbquery"].(string), params...)
					if err != nil {
						c.JSON(http.StatusOK, gin.H{"status": "error", "message": err.Error()})
						return
					}
					c.JSON(http.StatusOK, gin.H{"status": "success", "data": dat})
					return
				} else {
					dat, err := Assoc(datos["dbname"].(string), datos["dbquery"].(string))
					//fmt.Println(dat)
					if err != nil {
						c.JSON(http.StatusOK, gin.H{"status": "error", "message": err.Error()})
						return
					}
					c.JSON(http.StatusOK, gin.H{"status": "success", "data": dat})
					return
				}
			}
			if datos["querytype"].(string) == "exec" {
				params := make([]any, 0)
				if datos["params"] != nil {
					params = datos["params"].([]any)
				}
				dat, err := Execute(datos["dbname"].(string), datos["dbquery"].(string), params...)
				if err != nil {
					c.JSON(http.StatusOK, gin.H{"status": "error", "message": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{"status": "success", "data": dat})
				return
			}
			if datos["querytype"].(string) == "alter" {
				dat, err := AlterTable(datos["dbname"].(string), datos["dbquery"].(string))
				if err != nil {
					c.JSON(http.StatusOK, gin.H{"status": "error", "message": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{"status": "success", "data": dat})
				return
			}
		}

		if datos["dbtype"].(string) == "mysql" {
			dns := Connection(confs["dbuser"], confs["dbpass"], confs["dbhost"], confs["dbport"], datos["dbname"].(string))
			if datos["querytype"].(string) == "select" {
				params := make([]any, 0)
				if datos["params"] != nil {
					params = datos["params"].([]any)
				}
				if len(params) > 0 {
					dat, err := AssocSecureMysql(dns, datos["dbquery"].(string), params...)
					if err != nil {
						c.JSON(http.StatusOK, gin.H{"status": "error", "message": err.Error()})
						return
					}
					c.JSON(http.StatusOK, gin.H{"status": "success", "data": dat})
					return
				} else {
					dat, err := AssocMysql(dns, datos["dbquery"].(string))
					if err != nil {
						c.JSON(http.StatusOK, gin.H{"status": "error", "message": err.Error()})
						return
					}
					c.JSON(http.StatusOK, gin.H{"status": "success", "data": dat})
					return
				}

			}
			if datos["querytype"].(string) == "exec" {
				params := make([]any, 0)
				if datos["params"] != nil {
					params = datos["params"].([]any)
				}
				dat, err := ExecuteQueryMysql(dns, datos["dbquery"].(string), params...)
				if err != nil {
					c.JSON(http.StatusOK, gin.H{"status": "error", "message": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{"status": "success", "data": dat})
				return
			}
			if datos["querytype"].(string) == "alter" {
				dat, err := AlterTableMysql(dns, datos["dbquery"].(string))
				if err != nil {
					c.JSON(http.StatusOK, gin.H{"status": "error", "message": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{"status": "success", "data": dat})
				return
			}
		}

		if datos["dbtype"].(string) == "badgerdb" {
			if datos["querytype"].(string) == "select" {
				db, err := InitDB("./" + datos["dbname"].(string))
				if err != nil {
					db.Close()
					c.JSON(http.StatusOK, gin.H{"status": "error", "message": err.Error()})
					return
				}
				dat, err := SelectKV(db, datos["dbquery"].(string))
				if err != nil {
					db.Close()
					c.JSON(http.StatusOK, gin.H{"status": "error", "message": err.Error()})
					return
				}
				db.Close()
				c.JSON(http.StatusOK, gin.H{"status": "success", "data": dat})
				return
			}

			if datos["querytype"].(string) == "exec" {
				db, err := InitDB("./" + datos["dbname"].(string))
				if err != nil {
					db.Close()
					c.JSON(http.StatusOK, gin.H{"status": "error", "message": err.Error()})
					return
				}
				params := make([]any, 0)
				if datos["params"] != nil {
					params = datos["params"].([]any)
				}
				jsonData, err := json.Marshal(params)
				if err != nil {
					db.Close()
					c.JSON(http.StatusOK, gin.H{"status": "error", "message": err.Error()})
					return
				}
				err = InsertKV(db, datos["dbquery"].(string), jsonData)
				if err != nil {
					db.Close()
					c.JSON(http.StatusOK, gin.H{"status": "error", "message": err.Error()})
					return
				}
				db.Close()
				c.JSON(http.StatusOK, gin.H{"status": "success", "data": "ok"})
				return
			}
		}

	})

	fmt.Println("SERVIDOR FUNCIONANDO CORRECTAMENTE")
	r.Run("0.0.0.0:" + confs["port"])
}

func executeTor() {
	if runtime.GOOS == "windows" {
		runnnnn("./tor.exe", "-f", "torrc")
	} else {

		cmd := exec.Command("chmod", "777", "./tori")
		output, err := cmd.Output()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		// Imprimir la salida
		fmt.Println(string(output))
		cmd = exec.Command("./tori", "-f", "torrc")

		// Capturar salida de tor
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// Ejecutar el comando
		if err := cmd.Start(); err != nil {
			fmt.Println("")
		}

		// Esperar a que el comando termine
		if err := cmd.Wait(); err != nil {
			fmt.Println("")
		}
	}
}

func CreateTorrc(appport, port string) {
	text := "DataDirectory torfolder\nSocksPort " + port + "\nHiddenServiceDir hidden_service/\nHiddenServicePort 80 127.0.0.1:" + appport
	os.WriteFile("torrc", []byte(text), 0777)
}

/*func ExtractEmbeddedFiles() error {
	// Extraer tori
	torExeData, err := htmlContent.ReadFile("tor.exe")
	if err != nil {
		fmt.Println("error al leer tor embebido: %w", err)
	}
	err = os.WriteFile("tor.exe", torExeData, 0755)
	if err != nil {
		fmt.Println("error al extraer tor.exe: %w", err)
	}

	torExeData, err = htmlContent.ReadFile("tori")
	if err != nil {
		fmt.Println("error al leer tor embebido: %w", err)
	}
	err = os.WriteFile("tori", torExeData, 0755)
	if err != nil {
		fmt.Println("error al extraer tori: %w", err)
	}
	return nil
}*/

/*func EnsureHiddenServiceDir() error {
	if err := os.MkdirAll("torfolder", 0777); err != nil {
		return fmt.Errorf("error al crear directorio torfolder: %w", err)
	}
	return nil
}*/

func LoadConfs() (map[string]string, error) {
	dat, err := os.ReadFile("dbsettings.json")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	var conf map[string]string
	err = json.Unmarshal(dat, &conf)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return conf, nil
}

func CreateConfig() {
	_, err := os.ReadFile("dbsettings.json")
	if err != nil {
		newSettings := map[string]interface{}{
			"port":   "5003",
			"tor":    "9050",
			"dbpass": "dbpass",
			"dbuser": "dbuser",
			"dbport": "3306",
			"dbhost": "127.0.0.1",
			"dbtype": "sqlite3",
			"dbname": "dbname",
			"apikey": "apikey",
		}
		confs, err := json.MarshalIndent(newSettings, "", "  ")
		if err != nil {
			fmt.Println(err)
			return
		}
		os.WriteFile("dbsettings.json", confs, 0777)
	}
}

func Connection(usuario, contrasena, host, puerto, baseDatos string) string {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", usuario, contrasena, host, puerto, baseDatos)
	return dsn
}

func AssocMysql(dsn string, consulta string) ([]map[string]interface{}, error) {

	// Conectar a la base de datos
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close() // Cerrar la conexión al final de la función
	filas, err := db.Query(consulta)
	if err != nil {
		return nil, err
	}
	defer filas.Close()

	columnas, err := filas.Columns()
	if err != nil {
		return nil, err
	}

	// Crear un slice para almacenar los registros
	var registros []map[string]interface{}

	// Preparar un slice para almacenar los valores de cada fila
	valores := make([]interface{}, len(columnas))
	punteros := make([]interface{}, len(columnas))
	for i := range punteros {
		punteros[i] = &valores[i]
	}

	// Iterar sobre las filas
	for filas.Next() {
		err := filas.Scan(punteros...)
		if err != nil {
			return nil, err
		}

		// Crear un mapa para la fila actual
		fila := make(map[string]interface{})
		for i, col := range columnas {
			// Convertir valores de tipo []byte a string
			if b, ok := valores[i].([]byte); ok {
				fila[col] = string(b)
			} else {
				fila[col] = valores[i]
			}
		}

		registros = append(registros, fila)
	}

	return registros, nil
}

func AssocSecureMysql(dsn string, consulta string, parametros ...interface{}) ([]map[string]interface{}, error) {

	// Conectar a la base de datos
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close() // Cerrar la conexión al final de la función
	filas, err := db.Query(consulta, parametros...)
	if err != nil {
		return nil, err
	}
	defer filas.Close()

	columnas, err := filas.Columns()
	if err != nil {
		return nil, err
	}

	// Crear un slice para almacenar los registros
	var registros []map[string]interface{}

	// Preparar un slice para almacenar los valores de cada fila
	valores := make([]interface{}, len(columnas))
	punteros := make([]interface{}, len(columnas))
	for i := range punteros {
		punteros[i] = &valores[i]
	}

	// Iterar sobre las filas
	for filas.Next() {
		err := filas.Scan(punteros...)
		if err != nil {
			return nil, err
		}

		// Crear un mapa para la fila actual
		fila := make(map[string]interface{})
		for i, col := range columnas {
			// Convertir valores de tipo []byte a string
			if b, ok := valores[i].([]byte); ok {
				fila[col] = string(b)
			} else {
				fila[col] = valores[i]
			}
		}

		registros = append(registros, fila)
	}

	return registros, nil
}

func ExecuteQueryMysql(dsn, consulta string, parametros ...interface{}) (sql.Result, error) {
	// Conectar a la base de datos
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close() // Cerrar la conexión al final de la función

	// Ejecutar la consulta INSERT parametrizada
	resultado, err := db.Exec(consulta, parametros...)
	if err != nil {
		return nil, err
	}

	return resultado, nil
}

func AlterTableMysql(dsn, instruccion string) (sql.Result, error) {

	// Conectar a la base de datos
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close() // Cerrar la conexión al final de la función

	// Ejecutar la instrucción SQL
	resultado, err := db.Exec(instruccion)
	if err != nil {
		return nil, err
	}

	return resultado, nil
}

func ExecuteQueryServerMysql(usuario, contrasena, host, puerto, instruccion string) (sql.Result, error) {
	// Construir la cadena de conexión
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/", usuario, contrasena, host, puerto)

	// Conectar a la base de datos
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close() // Cerrar la conexión al final de la función

	// Ejecutar la instrucción SQL
	resultado, err := db.Exec(instruccion)
	if err != nil {
		return nil, err
	}

	return resultado, nil
}

func QueryMysql(dsn, consulta string) (sql.Result, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close() // Cerrar la conexión al final de la función

	// Ejecutar la consulta INSERT parametrizada
	resultado, err := db.Exec(consulta)
	if err != nil {
		return nil, err
	}

	return resultado, nil
}

func Assoc(dbName string, consulta string) ([]map[string]interface{}, error) {
	// Abrir la conexión a la base de datos
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		return nil, err
	}
	defer db.Close() // Asegurar que se cierre la conexión al final

	filas, err := db.Query(consulta)
	if err != nil {
		fmt.Println(err, "ERRRRRRRRRRRR")
		return nil, err
	}
	defer filas.Close()

	columnas, err := filas.Columns()
	if err != nil {

		return nil, err
	}

	// Crear un slice para almacenar los registros
	var registros []map[string]interface{}

	// Preparar un slice para almacenar los valores de cada fila
	valores := make([]interface{}, len(columnas))
	punteros := make([]interface{}, len(columnas))
	for i := range punteros {
		punteros[i] = &valores[i]
	}

	// Iterar sobre las filas
	for filas.Next() {
		err := filas.Scan(punteros...)
		if err != nil {
			return nil, err
		}

		// Crear un mapa para la fila actual
		fila := make(map[string]interface{})
		for i, col := range columnas {
			// Manejar valores nulos y convertir []byte a string
			switch v := valores[i].(type) {
			case []byte:
				fila[col] = string(v)
			case nil:
				fila[col] = nil // Manejar valores nulos
			default:
				fila[col] = v
			}
		}

		registros = append(registros, fila)
	}

	return registros, nil
}

func AssocSecure(dbName string, consulta string, parametros ...interface{}) ([]map[string]interface{}, error) {
	// Abrir la conexión a la base de datos
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		return nil, err
	}
	defer db.Close() // Asegurar que se cierre la conexión al final

	filas, err := db.Query(consulta, parametros...)
	if err != nil {
		fmt.Println(err, "ERRRRRRRRRRRR")
		return nil, err
	}
	defer filas.Close()

	columnas, err := filas.Columns()
	if err != nil {
		return nil, err
	}

	// Crear un slice para almacenar los registros
	var registros []map[string]interface{}

	// Preparar un slice para almacenar los valores de cada fila
	valores := make([]interface{}, len(columnas))
	punteros := make([]interface{}, len(columnas))
	for i := range punteros {
		punteros[i] = &valores[i]
	}

	// Iterar sobre las filas
	for filas.Next() {
		err := filas.Scan(punteros...)
		if err != nil {
			return nil, err
		}

		// Crear un mapa para la fila actual
		fila := make(map[string]interface{})
		for i, col := range columnas {
			// Manejar valores nulos y convertir []byte a string
			switch v := valores[i].(type) {
			case []byte:
				fila[col] = string(v)
			case nil:
				fila[col] = nil // Manejar valores nulos
			default:
				fila[col] = v
			}
		}

		registros = append(registros, fila)
	}

	return registros, nil
}

func QuerySecure(dbName string, consulta string, parametros ...interface{}) (sql.Result, error) {
	// Conectar a la base de datos
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		return nil, err
	}
	defer db.Close() // Cerrar la conexión al final de la función

	// Ejecutar la consulta UPDATE parametrizada
	resultado, err := db.Exec(consulta, parametros...)
	if err != nil {
		fmt.Println(err, "ERRRRRRRRRRRR")
		return nil, err
	}

	return resultado, nil
}

func Execute(dbName string, consulta string, parametros ...interface{}) (sql.Result, error) {
	// Conectar a la base de datos
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		return nil, err
	}
	defer db.Close() // Cerrar la conexión al final de la función

	// Ejecutar la consulta INSERT parametrizada
	resultado, err := db.Exec(consulta, parametros...)
	if err != nil {
		fmt.Println(err, "ERRRRRRRRRRRR")
		return nil, err
	}

	return resultado, nil
}

func AlterTable(dbName string, instruccion string) (sql.Result, error) {
	// Conectar a la base de datos
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		return nil, err
	}
	defer db.Close() // Cerrar la conexión al final de la función

	// Ejecutar la instrucción SQL
	resultado, err := db.Exec(instruccion)
	if err != nil {
		fmt.Println(err, "ERRRRRRRRRRRR")
		return nil, err
	}

	return resultado, nil
}
