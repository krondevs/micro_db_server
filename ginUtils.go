package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupLogger() {
	// Crear o abrir el archivo de logs
	file, err := os.OpenFile("error.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Print(err)
	}

	// Configurar el logger para escribir en el archivo
	log.SetOutput(file)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func errorLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Procesar la solicitud
		c.Next()

		// Comprobar si hubo errores
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				log.Println(err.Error()) // Registrar el error en el archivo
			}
		}
	}
}

func GinRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode) // Modo Release para despliegue
	r := gin.Default()
	r.Use(errorLogger())

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Permitir todos los orígenes
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		AllowCredentials: true,
	}))

	return r
}
