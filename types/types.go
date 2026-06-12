package types

import "time"

// ***************important*************
// I have to handle case of file move

type Create struct{
	Path string `json:"path"`
	Name string `json:"name"`
	Action string `json:"action"`
	IsDir bool `json:"isDir"`
	Size int64 `json:"size"`
	CreateTime time.Time `json:"createTime"`
}

type Remove struct{
	Path string `json:"path"`
	Name string `json:"name"`
	Action string `json:"action"`
	IsDir bool `json:"isDir"`
	Size int64 `json:"size"`
	CreateTime time.Time `json:"createTime"`
	RemoveTime time.Time `json:"removeTime"`
}

type Rename struct{
	OldPath string `json:"oldPath"`
	OldName string `json:"oldName"`
	Action string `json:"action"`
	IsDir bool `json:"isDir"`
	Size int64 `json:"size"`
	Path string `json:"path"`
	Name string `json:"name"`
	RenameTime time.Time `json:"createTime"`
}

type Write struct{
	Path      string    `json:"path"`  // file path
	Type      string    `json:"type,omitempty"`       // e.g. "snapshot", "delta"
	Content   string    `json:"content,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	CurrentSize int64  `json:"currentSize,omitempty"`
	PrevSize    int64  `json:"prevSize,omitempty"`
	PreviousFileContent string `json:"previousFileContent,omitempty"`
}

type Repository struct{
	UserName string `json:"username"`
	RemoteUrl string `json:"remoteUrl"`
}

type Recorder struct{
	DebounceTime int16  // in sec
	CodeThreshold int16  // in sec
}

type SSHKeys struct{
	PublicKeyPath string
	PrivateKeyPath string
}

type Config struct {
	WorkingDir string
	Repository Repository
	Recorder Recorder
	SSHKeys SSHKeys
}

type Node struct {
    Name     string     `json:"name"`
    Path     string     `json:"path"`               // absolute or repo-relative
    IsDir    bool       `json:"isDir"`
    Size     int64      `json:"size,omitempty"`     // bytes; 0 for dirs
    Children []*Node    `json:"children,omitempty"` // nil when no children -> omitted in JSON
}

type FileTree struct{
	Files []*Node `json:"files"`
}

type Metadata struct{
	UserName string `json:"userName"`
	RepoName string `json:"repoName"`
}