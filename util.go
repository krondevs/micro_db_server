package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

var Type = reflect.TypeOf
var Print = fmt.Println

const (
	minSliceSize     = 1000 // Tamaño mínimo para búsqueda paralela
	minPartitionSize = 250  // Tamaño mínimo por goroutine
)

func StringToHexa(s string) string {
	return hex.EncodeToString([]byte(s))
}

func HexaToString(hexStr string) (string, error) {
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func NumToInt64(i interface{}) int64 {
	switch v := i.(type) {
	case float64:
		return int64(v)
	case float32:
		return int64(v)
	case int:
		return int64(v)
	case int8:
		return int64(v)
	case int16:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case uint:
		return int64(v)
	case uint8:
		return int64(v)
	case uint16:
		return int64(v)
	case uint32:
		return int64(v)
	case uint64:
		return int64(v)
	case string:
		if num, err := strconv.ParseInt(v, 10, 64); err == nil {
			return num
		}
		return 0
	default:
		return 0
	}
}

func IsInMapStruct(m map[string]struct{}, key string) bool {
	_, exists := m[key]
	return exists
}

func IsInMapInterface(m map[string]interface{}, key string) bool {
	_, exists := m[key]
	return exists
}

func Trim(text string) string {
	gg := strings.TrimSpace(text)
	return gg
}

func EscapeSQL(value string) string {
	gg := strings.ReplaceAll(value, "'", "''")
	return gg
}

func GetRandom(bajo, alto int) int {
	return rand.Intn(alto-bajo) + bajo // Generar el número aleatorio
}

func GetValues(m map[string]interface{}) []interface{} {
	values := reflect.ValueOf(m)
	vals := make([]interface{}, values.Len())

	for i, key := range values.MapKeys() {
		vals[i] = values.MapIndex(key).Interface()
	}

	return vals
}

func GetKeys(m map[string]interface{}) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func ReplaceChar(input string, oldChar rune, newChar string) string {
	// Si newChar es un string vacío, se reemplazará oldChar por una cadena vacía
	if newChar == "" {
		return strings.ReplaceAll(input, string(oldChar), "")
	}
	return strings.ReplaceAll(input, string(oldChar), newChar)
}

func GetStructKeys(data map[string]struct{}) []string {
	var keys []string

	// Iterar sobre las claves del mapa
	for key := range data {
		keys = append(keys, key)
	}

	return keys
}

func StringToNum(valor string) (interface{}, error) {
	// Intentar convertir a entero
	if numeroInt, err := strconv.Atoi(valor); err == nil {
		return numeroInt, nil
	}

	// Intentar convertir a flotante
	if numeroFloat, err := strconv.ParseFloat(valor, 64); err == nil {
		return numeroFloat, nil
	}

	// Si ninguna conversión es exitosa, devolver un error
	return nil, fmt.Errorf("no se pudo convertir la cadena: %s", valor)
}

func NumToString(numero interface{}) string {
	switch v := numero.(type) {
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case float64:
		return strconv.FormatFloat(v, 'f', 2, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', 2, 32)
	default:
		return "0"
	}
}

func Str(numero interface{}) string {
	return NumToString(numero)
}

func Int(number interface{}) int64 {
	return NumToInt64(number)
}

func Float(number interface{}) float64 {
	return NumToFloat64(number)
}

func In(m map[string]struct{}, key string) bool {
	return IsInMapStruct(m, key)
}

func NumToFloat64(i interface{}) float64 {
	switch v := i.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int8:
		return float64(v)
	case int16:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case uint:
		return float64(v)
	case uint8:
		return float64(v)
	case uint16:
		return float64(v)
	case uint32:
		return float64(v)
	case uint64:
		return float64(v)
	case string:
		if num, err := strconv.ParseFloat(v, 64); err == nil {
			return num
		}
		return 0
	default:
		return 0
	}
}

/*
	func PrintRed(text ...string) {
		fondoRojo := color.New(color.FgWhite, color.BgRed)
		resultado := strings.Join(text, " ")
		fondoRojo.Println(resultado)
	}

	func PrintGreen(text ...string) {
		fondoVerde := color.New(color.FgBlack, color.BgGreen)
		resultado := strings.Join(text, " ")
		fondoVerde.Println(resultado)
	}

	func PrintYellow(text ...string) {
		fondoAmarillo := color.New(color.FgBlack, color.BgYellow)
		resultado := strings.Join(text, " ")
		fondoAmarillo.Println(resultado)
	}

	func PrintBlue(text ...string) {
		fondoAzul := color.New(color.FgWhite, color.BgBlue)
		resultado := strings.Join(text, " ")
		fondoAzul.Println(resultado)
	}
*/
func Input(prompt string) string {
	fmt.Print(prompt)                     // Imprimir el mensaje de aviso
	scanner := bufio.NewScanner(os.Stdin) // Crear un nuevo scanner
	scanner.Scan()                        // Leer la entrada del usuario
	return scanner.Text()                 // Devolver el texto ingresado
}

func Pause() {
	fmt.Print("Presiona Enter para salir...")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan() // Espera a que el usuario presione Enter
}

func ToJSON(data interface{}) (string, error) {
	// Convertir a JSON con sangrías
	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// export
func FromJSON(jsonStr string) (interface{}, error) {
	var data interface{}
	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		return nil, fmt.Errorf("error al convertir de JSON: %v", err)
	}
	return data, nil
}

func ParseFloat(value string) float64 {
	result, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0.0 // Maneja el error según sea necesario
	}
	return result
}

func ParseInt(value string) int64 {
	result, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0 // Maneja el error según sea necesario
	}
	return result
}

/*
func Format(cantidad float64, decimales int) string {
	decim := int64(cantidad)
	ss := cantidad - float64(decim)
	if ss != 0 {
		format := "%." + strconv.Itoa(decimales) + "f"
		return fmt.Sprintf(format, cantidad)
	}
	format := "%.0f"
	return fmt.Sprintf(format, cantidad)
}
*/

func Normalize(base, comparar, tolerancia float64) float64 {
	diferencia := comparar - base
	if diferencia <= tolerancia && diferencia >= -tolerancia {
		return base // Devuelve el valor base si está dentro de la tolerancia
	}
	return comparar // Devuelve el valor a comparar si no está dentro de la tolerancia
}

func CreateTorProxy(port string) (*http.Client, error) {
	proxyURL, err := url.Parse("socks5h://localhost:" + port)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
		Timeout: 30 * time.Second,
	}
	return client, nil
}

func PostDataTor(port, targetURL, data string) (string, error) {
	client, err := CreateTorProxy(port)
	if err != nil {
		return "", err
	}

	// Crear el cuerpo de la solicitud
	req, err := http.NewRequest("POST", targetURL, bytes.NewBuffer([]byte(data)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Realizar la solicitud POST
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Leer la respuesta
	var responseBody bytes.Buffer
	if _, err := responseBody.ReadFrom(resp.Body); err != nil {
		return "", err
	}

	return responseBody.String(), nil
}

func CreateHTML(template string, values []interface{}) string {
	templatem := fmt.Sprintf(template, values...)
	return templatem
}

func PostJson(url string, data map[string]interface{}) (map[string]interface{}, error) {
	// Convertir el mapa a JSON.
	dat, err := json.Marshal(data)
	if err != nil {
		//fmt.Println(err, "ERROR") // Imprimir el error si ocurre.
		return nil, err // Retornar false en caso de error.
	}

	// Crear una nueva solicitud HTTP POST con la URL y el cuerpo en formato JSON.
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(dat))
	if err != nil {
		//fmt.Println(err, "error 2") // Imprimir el error si ocurre.
		return nil, err // Retornar false en caso de error.
	}

	// Establecer el encabezado Content-Type para indicar que se está enviando un JSON.
	req.Header.Set("Content-Type", "application/json")

	// Crear un cliente HTTP para enviar la solicitud.
	client := &http.Client{}

	// Enviar la solicitud y capturar la respuesta.
	response, err := client.Do(req)
	if err != nil {
		//fmt.Println(err, "error 3") // Imprimir el error si ocurre.
		return nil, err // Retornar false en caso de error.
	}
	defer response.Body.Close() // Asegurarse de cerrar el cuerpo de la respuesta al finalizar.

	// Declarar un mapa para almacenar los datos de la respuesta del servidor.
	var serverData map[string]interface{}

	// Decodificar el cuerpo de la respuesta JSON en el mapa serverData.
	err = json.NewDecoder(response.Body).Decode(&serverData)
	if err != nil {
		//fmt.Println(err, "error 4") // Imprimir el error si ocurre.

		// Leer el cuerpo de la respuesta en caso de error de decodificación.
		body, _ := io.ReadAll(response.Body)
		fmt.Println(string(body)) // Imprimir el cuerpo de la respuesta para depuración.
		return nil, err           // Retornar false en caso de error.
	}
	return serverData, nil // Retornar true si la operación fue exitosa.

}

func HexDecode(hexStr string) string {
	decoded, err := hex.DecodeString(hexStr)
	if err != nil {
		return hexStr // Si hay error, devolver el string original
	}
	return string(decoded)
}

func gg(str string) string {
	return str
}

func PostUrlEncoded(url string, data map[string]interface{}) (map[string]interface{}, error) {
	// Construir la cadena URL encoded manualmente
	var parts []string

	for key, value := range data {
		// Convertir el valor a string hexadecimal
		var hexValue string
		switch v := value.(type) {
		case string:
			hexValue = hex.EncodeToString([]byte(v))
		default:
			// Para otros tipos, primero convertir a string
			str := fmt.Sprintf("%v", v)
			hexValue = hex.EncodeToString([]byte(str))
		}

		// Agregar el par clave=valor al slice
		parts = append(parts, key+"="+hexValue)
	}

	// Unir todas las partes con &
	encodedData := strings.Join(parts, "&")

	// Crear una nueva solicitud HTTP POST
	req, err := http.NewRequest("POST", url, strings.NewReader(encodedData))
	if err != nil {
		return nil, err
	}

	// Establecer el encabezado Content-Type
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Crear un cliente HTTP y enviar la solicitud
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// Procesar la respuesta
	var serverData map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&serverData)
	if err != nil {
		body, _ := io.ReadAll(response.Body)
		fmt.Println(string(body))
		return nil, err
	}
	return serverData, nil
}

func FormData(data map[string]string) *strings.Reader {
	var formDatos []string
	for key, value := range data {
		formDatos = append(formDatos, fmt.Sprintf("%s=%s", key, value))
	}
	return strings.NewReader(strings.Join(formDatos, "&"))
}

func RunServer(port string) {
	fmt.Println("Server running at ", "http://0.0.0.0:"+port)
	err := http.ListenAndServe("0.0.0.0:"+port, nil)
	if err != nil {
		panic(err)
	}
}

func ServeHTML() {
	// Manejador personalizado para servir archivos estáticos
	http.Handle("/static/", http.StripPrefix("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filePath := filepath.Join("static", r.URL.Path[len("/static/"):])
		if r.Method != http.MethodGet {
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
			return
		}
		http.ServeFile(w, r, filePath)

		// Establecer el tipo MIME
		ext := filepath.Ext(filePath)
		switch ext {
		case ".js":
			w.Header().Set("Content-Type", "application/javascript")
		case ".css":
			w.Header().Set("Content-Type", "text/css")
			// Agrega más tipos MIME según sea necesario
		}
	})))

	// Manejar la ruta raíz (/) y servir el archivo index.html desde la carpeta templates
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		host, _, err := net.SplitHostPort(ip)
		if err != nil {
			fmt.Println("Error al obtener la IP:", err)
			return
		}

		// Convertir a IPv4 si es necesario
		parsedIP := net.ParseIP(host)
		if parsedIP != nil && parsedIP.To4() != nil {
			fmt.Println("Conectado desde IP:", parsedIP.To4().String())
		} else {
			fmt.Println("Conectado desde IP:", host) // IPv6 o no se puede convertir
		}
		if ValidateMethod(r, "GET") {
			http.ServeFile(w, r, "templates/index.html")
		}
	})
}

func ValidateMethod(r *http.Request, method string) bool {
	if method == "POST" {
		return r.Method == http.MethodPost
	}
	if method == "GET" {
		return r.Method == http.MethodGet
	}
	if method == "PUT" {
		return r.Method == http.MethodPut
	}
	if method == "DELETE" {
		return r.Method == http.MethodDelete
	}
	if method == "PATCH" {
		return r.Method == http.MethodPatch
	}
	if method == "OPTIONS" {
		return r.Method == http.MethodOptions
	}
	if method == "HEAD" {
		return r.Method == http.MethodHead
	}
	return false
}

func RString(w http.ResponseWriter, content string, tipo string) {
	if tipo == "html" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(content))
	} else {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(content))
	}
}

func RJson(w http.ResponseWriter, content interface{}) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(content)
}

func RegisterTypes() {
	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})
}

func ReadJsonDatabase(fileName string) (map[string]interface{}, error) {
	_, err := os.Stat(fileName + ".lock")
	if !os.IsNotExist(err) {
		return nil, err
	}
	datosRecuperados, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	var datosLeidos map[string]interface{}
	err = json.Unmarshal(datosRecuperados, &datosLeidos)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	//fmt.Println(datosLeidos)
	return datosLeidos, nil
}

func UpdateJsonDatabase(fileName string, data map[string]interface{}) error {
	lockFile, err := os.OpenFile(fileName+".lock", os.O_CREATE|os.O_EXCL|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	lockFile.Close()
	defer os.Remove(fileName + ".lock")
	datosJson, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println(err)
		return err
	}
	err = os.WriteFile(fileName, datosJson, 0644)
	if err != nil {
		return err
	}
	return nil
}

func UpdateFile(fileName string, data []byte) error {
	lockFile, err := os.OpenFile(fileName+".lock", os.O_CREATE|os.O_EXCL|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	lockFile.Close()
	defer os.Remove(fileName + ".lock")
	err = os.WriteFile(fileName, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

// ReadGobDatabase lee un archivo en formato Gob y devuelve un mapa de interfaces.
func ReadGobDatabase(fileName string) (map[string]interface{}, error) {
	_, err := os.Stat(fileName + ".lock")
	if !os.IsNotExist(err) {
		return nil, err
	}

	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer file.Close()

	var datosLeidos map[string]interface{}
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&datosLeidos)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	fmt.Println(datosLeidos)
	return datosLeidos, nil
}

// UpdateGobDatabase actualiza un archivo en formato Gob con los datos proporcionados.
func UpdateGobDatabase(fileName string, data map[string]interface{}) error {
	RegisterTypes()
	lockFile, err := os.OpenFile(fileName+".lock", os.O_CREATE|os.O_EXCL|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	lockFile.Close()
	defer os.Remove(fileName + ".lock")
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(data)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func Hoy() string {
	return time.Now().Format("2006-01-02")
}

func DateTime() string {
	return time.Now().Format("2006-01-02 15:04:05") // Formato de Go
}

func HourNow() string {
	return time.Now().Format("15:04:05") // Formato de Go
}

func UnixTime() int64 {
	return time.Now().Unix() // Devuelve el tiempo en segundos desde epoch
}

func UnixMillisecTime() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond) // Devuelve el tiempo en milisegundos desde epoch
}

func UnixToDate(epoch int64) string {
	// Convertir el epoch a un objeto Time
	t := time.Unix(epoch, 0)
	// Formatear la fecha y hora
	return t.Format("2006-01-02")
}

func DiaHoy() string {
	fechaActual := time.Now()
	nombreDia := fechaActual.Weekday().String()

	diasSemana := map[string]string{
		"Monday":    "Lunes",
		"Tuesday":   "Martes",
		"Wednesday": "Miércoles",
		"Thursday":  "Jueves",
		"Friday":    "Viernes",
		"Saturday":  "Sábado",
		"Sunday":    "Domingo",
	}

	return diasSemana[nombreDia]
}

func NodeRequest(method string, params []interface{}) (map[string]interface{}, error) {
	confs, err := LeerConfiguracionNode("Kron.conf")
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("http://127.0.0.1:%s/", confs["rpcport"])
	headers := map[string]string{"content-type": "text/plain"}

	data := map[string]interface{}{
		"jsonrpc": "1.0",
		"id":      "curltest",
		"method":  method,
		"params":  params,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Autenticación básica
	req.SetBasicAuth(confs["rpcuser"], confs["rpcpassword"])

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

type Config struct {
	RpcUser     string
	RpcPassword string
	RpcPort     string
}

func LeerConfiguracionNode(filename string) (map[string]string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	inString := string(content)
	configuracion := make(map[string]string)
	splitedText := strings.Split(inString, "\n")
	var subSlice []string
	for _, i := range splitedText {
		subSlice = strings.Split(i, "=")
		if len(subSlice) > 1 {
			//fmt.Println(subSlice)
			configuracion[subSlice[0]] = subSlice[1]
		}
	}
	return configuracion, nil

}

func NodeRequestKron(method string, params []interface{}) (interface{}, error) {
	confs, err := LeerConfiguracionNode("Kron.conf")
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("http://127.0.0.1:%s/", confs["rpcport"])
	headers := map[string]string{"content-type": "text/plain"}

	data := map[string]interface{}{
		"jsonrpc": "1.0",
		"id":      "curltest",
		"method":  method,
		"params":  params,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Autenticación básica
	req.SetBasicAuth(confs["rpcuser"], confs["rpcpassword"])

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, err
	}
	fmt.Println(result)

	return result["result"], nil
}

func GetAssetBalance(address string, asset string) float64 {
	dat, err := NodeRequestKron("listassetbalancesbyaddress", []any{address})
	if err != nil {
		fmt.Println(err)
		return 0
	}
	if dat == nil {
		return 0
	}
	var saldo float64 = 0
	balances := dat.(map[string]any)
	for i := range balances {
		if i == asset {
			saldo = NumToFloat64(balances[i])
			break
		}
	}
	return saldo
}

func GetHash(input string) string {
	// Crear un nuevo hash SHA-256
	hash := sha256.New()
	// Escribir el string en el hash
	hash.Write([]byte(input))
	// Obtener el resultado del hash
	return hex.EncodeToString(hash.Sum(nil))
}

func GetMasterAddress() string {
	dat, err := NodeRequestKron("getaddressesbyaccount", []any{"sistwallet"})
	fmt.Println(dat)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	dirs := dat.([]interface{})
	if len(dirs) > 0 {
		return dirs[0].(string)
	}
	dat, err = NodeRequestKron("getnewaddress", []any{"sistwallet"})
	if err != nil {
		fmt.Println(err)
		return ""
	}
	address := strings.TrimSpace(dat.(string))
	return address
}

func Truncate(f float64, decimals int) float64 {
	pow := math.Pow(10, float64(decimals))
	return math.Trunc(f*pow) / pow
}

func TransferAsset(fromaddress string, asset string, amount float64, message string, toaddress string) (string, error) {
	availableBalance := GetAssetBalance(fromaddress, asset)
	amount = Truncate(amount, 7)
	if availableBalance < amount {
		return "", errors.New("not enough balance")
	}
	if amount <= 0.000001 {
		return "", errors.New("amount is too low")
	}
	message = GetHash(message)
	sistwallet := GetMasterAddress()
	dat, err := NodeRequestKron("transferfromaddress", []any{asset, fromaddress, amount, toaddress, message, int64(4090368000), sistwallet, fromaddress})
	if err != nil {
		return "", err
	}
	fmt.Println(dat)
	retorno := dat.([]any)
	txid := retorno[0].(string)
	return txid, nil
}

func BinarySearch(slice []interface{}, elemento interface{}) bool {
	// Para slices pequeños, búsqueda secuencial directa
	if len(slice) < minSliceSize {
		for _, v := range slice {
			if v == elemento {
				return true
			}
		}
		return false
	}

	// Configuración para búsqueda paralela
	numCPU := runtime.NumCPU()
	numGoroutines := numCPU * 2
	partitionSize := len(slice) / numGoroutines

	// Ajustar número de goroutines si las particiones son muy pequeñas
	if partitionSize < minPartitionSize {
		numGoroutines = len(slice) / minPartitionSize
		if numGoroutines < 1 {
			numGoroutines = 1
		}
		partitionSize = len(slice) / numGoroutines
	}

	var encontrado int32
	done := make(chan struct{})

	// Iniciar búsqueda paralela
	for i := 0; i < numGoroutines; i++ {
		start := i * partitionSize
		end := start + partitionSize
		if i == numGoroutines-1 {
			end = len(slice)
		}

		go func(start, end int) {
			// Búsqueda en la partición asignada
			for j := start; j < end; j++ {
				// Si ya se encontró en otra goroutine, terminar
				if atomic.LoadInt32(&encontrado) == 1 {
					return
				}

				if slice[j] == elemento {
					atomic.StoreInt32(&encontrado, 1)
					close(done)
					return
				}
			}
		}(start, end)
	}

	// Esperar hasta encontrar el elemento o terminar la búsqueda
	select {
	case <-done:
		return true
	default:
		// Esperar un poco más por si alguna goroutine está por encontrar el elemento
		for atomic.LoadInt32(&encontrado) == 0 {
			runtime.Gosched()
		}
		return atomic.LoadInt32(&encontrado) == 1
	}
}

func RequestsPost(url string, urlEncodedParams string) (string, error) {
	dat, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(urlEncodedParams))
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer dat.Body.Close()
	body, _ := io.ReadAll(dat.Body)
	return string(body), nil
}

/*
 */

/*
import (
    "fmt"
    "net/http"
	"encoding/json"
)


// Handler para la ruta /productos
func productosHandler(w http.ResponseWriter, r *http.Request) {
    // Asegurarse de que la solicitud sea de tipo POST
    if r.Method != http.MethodPost {
        http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
        return
    }

    // Parsear los datos del formulario
    err := r.ParseForm()
    if err != nil {
        http.Error(w, "Error al parsear los datos del formulario", http.StatusBadRequest)
        return
    }

    // Obtener los datos del formulario
    nombreProducto := r.FormValue("nombre_producto")
    cantidad := r.FormValue("cantidad")
    precio := r.FormValue("precio")
	datos := []map[string]interface{} {
		{"nombre_producto": "harina de trigo", "cantidad": 5, "precio": 1.3},
		{"nombre_producto": "pasta larga", "cantidad": 10, "precio": 80},
		{"nombre_producto": "azúcar", "cantidad": 3, "precio": 0.5},
		{"nombre_producto": "arroz", "cantidad": 7, "precio": 1.2},
		{"nombre_producto": "sal", "cantidad": 2, "precio": 0.3},
		{"nombre_producto": "aceite", "cantidad": 4, "precio": 2.5},
		{"nombre_producto": "leche", "cantidad": 6, "precio": 1.1},
		{"nombre_producto": "huevo", "cantidad": 12, "precio": 0.2},
	}
	response := map[string]interface{}{
		"mensaje":"productos recibidos",
		"producto": map[string]interface{}{
				"nombre producto": nombreProducto,
				"cantidad": cantidad,
				"precio":precio,
			},
		"datos":datos,
	}
    // Responder con los datos recibidos
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}

func RString(w http.ResponseWriter, content string, tipo string){
	if tipo == "html"{
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(content))
	}else{
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(content))
	}
}

func RJson(w http.ResponseWriter, content interface{}){
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(content)
}

func usuarioHandler(w http.ResponseWriter, r *http.Request) {
    // Asegurarse de que la solicitud sea de tipo POST
    if r.Method != http.MethodPost {
        http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
        return
    }

    // Parsear los datos del formulario
    err := r.ParseForm()
    if err != nil {
        http.Error(w, "Error al parsear los datos del formulario", http.StatusBadRequest)
        return
    }

    // Obtener los datos del formulario
    nombre := r.FormValue("nombre")
    edad := r.FormValue("edad")
    email := r.FormValue("email")

    // Responder con los datos recibidos
    w.Header().Set("Content-Type", "application/json")
    response := fmt.Sprintf(`{"mensaje": "Usuario recibido", "usuario": {"nombre": "%s", "edad": "%s", "email": "%s"}}`, nombre, edad, email)
    w.WriteHeader(http.StatusOK)
    fmt.Fprint(w, response)
}
/*
func main() {
    // Registrar el handler para la ruta /productos
    http.HandleFunc("/productos", productosHandler)
	http.HandleFunc("/usuario", usuarioHandler)

    // Iniciar el servidor en el puerto 8080
    fmt.Println("Servidor escuchando en :8086")
    if err := http.ListenAndServe(":8086", nil); err != nil {
        fmt.Println("Error al iniciar el servidor:", err)
    }
}
*/
