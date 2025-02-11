//go:build windows

package main

import (
	"fmt"
	"os/exec"
	"syscall"
)

func runCommand(command string, args ...string) error {
	cmd := exec.Command(command, args...)

	// En Windows, oculta la ventana de consola del proceso hijo

	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	// No necesitamos capturar stdout/stderr ya que queremos ejecución silenciosa
	//cmd.Stdout = nil
	//cmd.Stderr = nil
	stdout, err := cmd.Output()
	if err != nil {
		return err
	}

	// Imprimir la salida
	fmt.Println(string(stdout))

	// Start() es correcto aquí ya que no necesitamos esperar a que termine
	return cmd.Start()

	//fmt.Println(command, args)
	//return nil
}

func runnnnn(command string, args ...string) error {
	cmd := exec.Command(command, args...)

	// En Windows, oculta la ventana de consola del proceso hijo
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	// No necesitamos capturar stdout/stderr ya que queremos ejecución silenciosa
	//cmd.Stdout = nil
	//cmd.Stderr = nil
	stdout, err := cmd.Output()
	if err != nil {
		return err
	}

	// Imprimir la salida
	fmt.Println(string(stdout))

	// Start() es correcto aquí ya que no necesitamos esperar a que termine
	return cmd.Start()
}
