package common

import (
	"os"
	"testing"

	"github.com/spf13/afero"
)

var testFilePath afero.File

func TestFileExists(t *testing.T) {
	// Mock Vault secret path
	appFS := afero.NewMemMapFs()
	appFS.MkdirAll("/tmp/testfile", 0755)
	afero.WriteFile(appFS, "/tmp/testfile", []byte("test value"), 0755)
	testFilePath, _ = appFS.OpenFile("/tmp/testfile", os.O_RDONLY, 0755)

	type args struct {
		fs       afero.Fs
		filename string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "If a file exists, should return true",
			args: args{
				fs:       appFS,
				filename: testFilePath.Name(),
			},
			want: true,
		},
		{
			name: "If a file does not exist, should return false",
			args: args{
				fs:       appFS,
				filename: "/some/file",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FileExists(tt.args.fs, tt.args.filename); got != tt.want {
				t.Errorf("FileExists() = %v, want %v", got, tt.want)
			}
		})
	}
}
