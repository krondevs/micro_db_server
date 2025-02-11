Copiar
# Microservidor API de Bases de Datos

## Descripción
API intermedia en Go para consultas seguras a bases de datos SQLite3, MariaDB y MySQL.

## Características
- Soporte para SQLite3, MariaDB y MySQL
- Autenticación mediante API Key
- Consultas parametrizadas
- Respuestas en formato JSON
- Protección contra inyección SQL

## Formato de Solicitud

```json
{
    "dbtype": "mysql",
    "dbname": "database_name",
    "apikey": "your_api_key",
    "querytype": "select",
    "dbquery": "SELECT * FROM table WHERE condition = ?",
    "params": ["value"]
}
Tipos de Consultas Soportadas
select
insert
update
delete
Respuesta Típica
Consulta Exitosa
Copiar
{
  "data": [],
  "status": "success"
}
Consulta con Resultados
Copiar
{
  "data": [
    {
      "id": 1,
      "nombre": "Producto",
      "precio": 100.50
    }
  ],
  "status": "success"
}
Respuesta de Error
Copiar
{
  "error": "Mensaje de error",
  "status": "error"
}
Ejemplo de Consulta
Consulta Simple
Copiar
{
    "dbtype": "mysql",
    "dbname": "tienda",
    "apikey": "secret_key",
    "querytype": "select",
    "dbquery": "SELECT * FROM productos",
    "params": []
}
Consulta Parametrizada
Copiar
{
    "dbtype": "sqlite3",
    "dbname": "local.db",
    "apikey": "secret_key",
    "querytype": "select", 
    "dbquery": "SELECT * FROM usuarios WHERE edad > ?",
    "params": [18]
}
Endpoint
/query
Método: POST
Autenticación: Requiere API Key
Content-Type: application/json
Códigos de Estado HTTP
Código	Descripción
200	Éxito
400	Solicitud incorrecta
401	No autorizado
403	Acceso prohibido
500	Error interno del servidor
Consideraciones de Seguridad
Usar siempre HTTPS
Proteger API Key
Utilizar consultas parametrizadas
Limitar permisos de base de datos
Bases de Datos Soportadas
SQLite3
MariaDB
MySQL
