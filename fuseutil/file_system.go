// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fuseutil

import (
	"io"

	"github.com/jacobsa/fuse"
	"github.com/jacobsa/fuse/fuseops"
)

// An interface with a method for each op type in the fuseops package. This can
// be used in conjunction with NewFileSystemServer to avoid writing a "dispatch
// loop" that switches on op types, instead receiving typed method calls
// directly.
//
// Each method is responsible for calling Respond on the supplied op.
//
// See NotImplementedFileSystem for a convenient way to embed default
// implementations for methods you don't care about.
type FileSystem interface {
	Init(*fuseops.InitOp)
	LookUpInode(*fuseops.LookUpInodeOp)
	GetInodeAttributes(*fuseops.GetInodeAttributesOp)
	SetInodeAttributes(*fuseops.SetInodeAttributesOp)
	ForgetInode(*fuseops.ForgetInodeOp)
	MkDir(*fuseops.MkDirOp)
	CreateFile(*fuseops.CreateFileOp)
	RmDir(*fuseops.RmDirOp)
	Unlink(*fuseops.UnlinkOp)
	OpenDir(*fuseops.OpenDirOp)
	ReadDir(*fuseops.ReadDirOp)
	ReleaseDirHandle(*fuseops.ReleaseDirHandleOp)
	OpenFile(*fuseops.OpenFileOp)
	ReadFile(*fuseops.ReadFileOp)
	WriteFile(*fuseops.WriteFileOp)
	SyncFile(*fuseops.SyncFileOp)
	FlushFile(*fuseops.FlushFileOp)
	ReleaseFileHandle(*fuseops.ReleaseFileHandleOp)
}

// Create a fuse.Server that handles ops by calling the associated FileSystem
// method.Respond with the resulting error. Unsupported ops are responded to
// directly with ENOSYS.
//
// FileSystem methods are called ine exactly the order of supported ops
// received by the connection, on a single goroutine. The methods should
// probably not block, instead continuing long-running operations in the
// background. Note however that there are subtleties here: for example, you
// probably want to serialize the order of write and flush operations.
func NewFileSystemServer(fs FileSystem) fuse.Server {
	return fileSystemServer{fs}
}

// A convenience function that makes it easy to ensure you respond to an
// operation when a FileSystem method returns. Responds to op with the current
// value of *err.
//
// For example:
//
//     func (fs *myFS) ReadFile(op *fuseops.ReadFileOp) {
//       var err error
//       defer fuseutil.RespondToOp(op, &err)
//
//       if err = fs.frobnicate(); err != nil {
//         err = fmt.Errorf("frobnicate: %v", err)
//         return
//       }
//
//       // Lots more manipulation of err, and return paths.
//       // [...]
//     }
//
func RespondToOp(op fuseops.Op, err *error) {
	op.Respond(*err)
}

type fileSystemServer struct {
	fs FileSystem
}

func (s fileSystemServer) ServeOps(c *fuse.Connection) {
	for {
		op, err := c.ReadOp()
		if err == io.EOF {
			break
		}

		if err != nil {
			panic(err)
		}

		switch typed := op.(type) {
		default:
			op.Respond(fuse.ENOSYS)

		case *fuseops.InitOp:
			s.fs.Init(typed)

		case *fuseops.LookUpInodeOp:
			s.fs.LookUpInode(typed)

		case *fuseops.GetInodeAttributesOp:
			s.fs.GetInodeAttributes(typed)

		case *fuseops.SetInodeAttributesOp:
			s.fs.SetInodeAttributes(typed)

		case *fuseops.ForgetInodeOp:
			s.fs.ForgetInode(typed)

		case *fuseops.MkDirOp:
			s.fs.MkDir(typed)

		case *fuseops.CreateFileOp:
			s.fs.CreateFile(typed)

		case *fuseops.RmDirOp:
			s.fs.RmDir(typed)

		case *fuseops.UnlinkOp:
			s.fs.Unlink(typed)

		case *fuseops.OpenDirOp:
			s.fs.OpenDir(typed)

		case *fuseops.ReadDirOp:
			s.fs.ReadDir(typed)

		case *fuseops.ReleaseDirHandleOp:
			s.fs.ReleaseDirHandle(typed)

		case *fuseops.OpenFileOp:
			s.fs.OpenFile(typed)

		case *fuseops.ReadFileOp:
			s.fs.ReadFile(typed)

		case *fuseops.WriteFileOp:
			s.fs.WriteFile(typed)

		case *fuseops.SyncFileOp:
			s.fs.SyncFile(typed)

		case *fuseops.FlushFileOp:
			s.fs.FlushFile(typed)

		case *fuseops.ReleaseFileHandleOp:
			s.fs.ReleaseFileHandle(typed)
		}
	}
}
