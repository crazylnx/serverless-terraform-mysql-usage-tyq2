package terraform

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path/filepath"
)

type execStruct struct {
	str string
	err error
}

func Execute(command string, vars []string, logger *logrus.Entry, stop chan int, secrets *[]string) (string, error) {

	execChan := make(chan execStruct)
	logger.Info(fmt.Sprintf("Begin to execute command: %s %s", command, vars))
	cmd := exec.Command(command, vars...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, *secrets...)

	if vars[0] == "init" {
		go terraformInit(execChan, cmd)
	} else {
		go terraformApplyAndDestroy(execChan, cmd, logger)
	}
	for {
		select {
		case execRes := <-execChan:
			if execRes.err != nil {
				return execRes.str, execRes.err
			}
			return execRes.str, nil
		// receive stop from pre freeze
		case <-stop:
			if err := cmd.Process.Signal(os.Interrupt); err != nil {
				return "", err
			}
		}
	}

}

func terraformApplyAndDestroy(execChan chan execStruct, cmd *exec.Cmd, logger *logrus.Entry) {
	path, err := filepath.Abs(".")
	if err != nil {
		execChan <- execStruct{
			str: "",
			err: err,
		}
		return
	}
	// all the output will be set in this buffer.
	outputBuf := new(bytes.Buffer)

	cmd.Dir = path + "/data/"
	output, err := cmd.StdoutPipe()
	if err != nil {
		execChan <- execStruct{
			str: "",
			err: err,
		}
		return
	}

	cmd.Stderr = cmd.Stdout
	if err := cmd.Start(); err != nil {
		execChan <- execStruct{
			str: "",
			err: err,
		}
		return
	}
	// stream outputBuf
	scanner := bufio.NewScanner(output)
	for scanner.Scan() {
		//wrapper := &OutputsWrapper{}
		m := scanner.Text()
		_, err = outputBuf.WriteString(m)
		if err != nil {
			execChan <- execStruct{
				str: "",
				err: err,
			}
			return
		}
		logger.Info(m)
	}

	if err = cmd.Wait(); err != nil {
		execChan <- execStruct{
			str: outputBuf.String(),
			err: err,
		}
		return
	}
	execChan <- execStruct{
		str: outputBuf.String(),
		err: nil,
	}
	return
}

// terraformInit because `terraform init` doesn't have a `-json` output, add this function to handle init.
// @todo, write more lines in one log.
func terraformInit(execChan chan execStruct, cmd *exec.Cmd) {
	path, err := filepath.Abs(".")
	if err != nil {
		execChan <- execStruct{
			str: "",
			err: err,
		}
		return
	}
	// all the output will be set in this buffer.
	outputBuf := new(bytes.Buffer)

	cmd.Dir = path + "/data/"
	output, err := cmd.StdoutPipe()

	if err != nil {
		execChan <- execStruct{
			str: "",
			err: err,
		}
		return
	}
	// combine two std.
	cmd.Stderr = cmd.Stdout
	if err := cmd.Start(); err != nil {
		execChan <- execStruct{
			str: "",
			err: err,
		}
		return
	}
	// stream outputBuf
	scanner := bufio.NewScanner(output)
	for scanner.Scan() {
		m := scanner.Text()
		_, err = outputBuf.WriteString(m)
		if err != nil {
			execChan <- execStruct{
				str: "",
				err: err,
			}
			return
		}
		fmt.Println(m)
	}

	if err = cmd.Wait(); err != nil {
		execChan <- execStruct{
			str: outputBuf.String(),
			err: err,
		}
		return
	}
	execChan <- execStruct{
		str: outputBuf.String(),
		err: nil,
	}
	return
}
