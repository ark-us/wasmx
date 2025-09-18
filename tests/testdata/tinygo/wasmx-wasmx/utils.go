package wasmxwasmx

import (
	"encoding/hex"
	"fmt"
	"path"
	"strings"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

var PINNED_FOLDER = "pinned"
var FILE_EXTENSIONS = map[string]string{
	wasmx.ROLE_INTERPRETER_PYTHON: "py",
	wasmx.ROLE_INTERPRETER_JS:     "js",
}

func WasmBuildPath(dataDir string, checksum []byte) string {
	return path.Join(dataDir, hex.EncodeToString(checksum))
}

func WasmBuildPathCompiled(dataDir string, checksum []byte) string {
	return path.Join(dataDir, PINNED_FOLDER, hex.EncodeToString(checksum))
}

func WasmBuildPathUtf8(dataDir string, checksum []byte, extension string) string {
	filename := fmt.Sprintf("%s_%s.%s", extension, hex.EncodeToString(checksum), extension)
	return path.Join(dataDir, extension, filename)
}

func GetPathUtf8(dataDir string, codeInfo wasmx.CodeInfo) string {
	return WasmBuildPathUtf8(dataDir, codeInfo.CodeHash, GetExtensionFromDeps(codeInfo.Deps))
}

func GetExtensionFromDeps(deps []string) string {
	extension := ""
	for _, dep := range deps {
		for role, ext := range FILE_EXTENSIONS {
			if strings.Contains(dep, role) {
				extension = ext
				break
			}
		}
	}
	return extension
}

func GetCodeFilePath(dataDir string, codeInfo wasmx.CodeInfo) string {
	filepath := ""
	if HasUtf8Dep(codeInfo.Deps) {
		filepath = GetPathUtf8(dataDir, codeInfo)
	} else {
		if len(codeInfo.InterpretedBytecodeRuntime) > 0 {
			filepath = ""
		} else {
			filepath = WasmBuildPath(dataDir, codeInfo.CodeHash)
		}
	}
	return filepath
}

func GetFilePath(dataDir string, codeInfo wasmx.CodeInfo) string {
	filepath := ""
	if codeInfo.Pinned {
		filepath = WasmBuildPathCompiled(dataDir, codeInfo.CodeHash)
	} else {
		if HasUtf8Dep(codeInfo.Deps) {
			filepath = GetPathUtf8(dataDir, codeInfo)
		} else {
			if len(codeInfo.InterpretedBytecodeRuntime) > 0 {
				filepath = ""
			} else {
				filepath = WasmBuildPath(dataDir, codeInfo.CodeHash)
			}
		}
	}
	return filepath
}

func HasUtf8Dep(deps []string) bool {
	for _, dep := range deps {
		if strings.Contains(dep, "utf8") {
			return true
		}
	}
	return false
}
