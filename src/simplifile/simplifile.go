package simplifile_P

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"syscall"

	gleam_P "example.com/todo/gleam"
)

const defaultFilePerm = 0o666
const defaultDirPerm = 0o777

func ReadBits(filepath gleam_P.String_t) gleam_P.Result_t[gleam_P.BitArray_t, FileError_t] {
	return gleamResult(func() (gleam_P.BitArray_t, error) {
		return os.ReadFile(path.Clean(string(filepath)))
	})
}

func WriteBits(filepath gleam_P.String_t, bits gleam_P.BitArray_t) gleam_P.Result_t[gleam_P.Nil_t, FileError_t] {
	return gleamResult(func() (gleam_P.Nil_t, error) {
		return gleam_P.Nil_c{}, os.WriteFile(path.Clean(string(filepath)), bits, defaultFilePerm)
	})
}

func AppendBits(filepath gleam_P.String_t, bits gleam_P.BitArray_t) gleam_P.Result_t[gleam_P.Nil_t, FileError_t] {
	return gleamResult(func() (gleam_P.Nil_t, error) {
		f, err := os.OpenFile(path.Clean(string(filepath)), os.O_CREATE|os.O_APPEND|os.O_WRONLY, defaultFilePerm)
		if err != nil {
			return gleam_P.Nil_c{}, err
		}
		defer f.Close()
		_, err = f.Write(bits)
		return gleam_P.Nil_c{}, err
	})
}

func IsFile(filepath gleam_P.String_t) gleam_P.Result_t[gleam_P.Bool_t, FileError_t] {
	return gleamResult(func() (gleam_P.Bool_t, error) {
		stat, err := os.Stat(path.Clean(string(filepath)))
		if err != nil {
			if os.IsNotExist(err) {
				return false, nil
			}
			return false, err
		}
		return gleam_P.Bool_t(stat.Mode().IsRegular()), nil
	})
}

func IsSymlink(filepath gleam_P.String_t) gleam_P.Result_t[gleam_P.Bool_t, FileError_t] {
	return gleamResult(func() (gleam_P.Bool_t, error) {
		stat, err := os.Lstat(path.Clean(string(filepath)))
		if err != nil {
			if os.IsNotExist(err) {
				return false, nil
			}
			return false, err
		}
		return stat.Mode()&os.ModeSymlink != 0, nil
	})
}

func IsDirectory(filepath gleam_P.String_t) gleam_P.Result_t[gleam_P.Bool_t, FileError_t] {
	return gleamResult(func() (gleam_P.Bool_t, error) {
		stat, err := os.Stat(path.Clean(string(filepath)))
		if err != nil {
			if os.IsNotExist(err) {
				return false, nil
			}
			return false, err
		}
		return gleam_P.Bool_t(stat.IsDir()), nil
	})
}

func CreateSymlink(target, symlink gleam_P.String_t) gleam_P.Result_t[gleam_P.Nil_t, FileError_t] {
	return gleamResult(func() (gleam_P.Nil_t, error) {
		return gleam_P.Nil_c{}, os.Symlink(string(target), string(symlink))
	})
}

func CreateDirectory(filepath gleam_P.String_t) gleam_P.Result_t[gleam_P.Nil_t, FileError_t] {
	return gleamResult(func() (gleam_P.Nil_t, error) {
		return gleam_P.Nil_c{}, os.Mkdir(path.Clean(string(filepath)), defaultDirPerm)
	})
}

func doCreateDirAll(dirpath gleam_P.String_t) gleam_P.Result_t[gleam_P.Nil_t, FileError_t] {
	return gleamResult(func() (gleam_P.Nil_t, error) {
		return gleam_P.Nil_c{}, os.MkdirAll(path.Clean(string(dirpath)), defaultDirPerm)
	})
}

func Delete(filepath gleam_P.String_t) gleam_P.Result_t[gleam_P.Nil_t, FileError_t] {
	return gleamResult(func() (gleam_P.Nil_t, error) {
		isDir := IsDirectory(filepath)
		if isDir.IsOk() && isDir.AsOk().P_0 {
			return gleam_P.Nil_c{}, os.RemoveAll(path.Clean(string(filepath)))
		} else {
			return gleam_P.Nil_c{}, os.Remove(path.Clean(string(filepath)))
		}
	})
}

func ReadDirectory(filepath gleam_P.String_t) gleam_P.Result_t[gleam_P.List_t[gleam_P.String_t], FileError_t] {
	return gleamResult(func() (gleam_P.List_t[gleam_P.String_t], error) {
		filepath := path.Clean(string(filepath))
		entries, err := os.ReadDir(filepath)
		if err != nil {
			return nil, err
		}
		var res []gleam_P.String_t
		for _, entry := range entries {
			res = append(res, gleam_P.String_t(path.Join(filepath, entry.Name())))
		}
		return gleam_P.ToList(res...), nil
	})
}

func doCopyFile(src, dest gleam_P.String_t) gleam_P.Result_t[gleam_P.Int_t, FileError_t] {
	return gleamResult(func() (gleam_P.Int_t, error) {
		srcFile, err := os.Open(path.Clean(string(src)))
		if err != nil {
			return 0, err
		}
		defer srcFile.Close()
		destFile, err := os.Create(path.Clean(string(dest)))
		if err != nil {
			return 0, err
		}
		defer destFile.Close()
		res, err := io.Copy(destFile, srcFile)
		return gleam_P.Int_t(res), err
	})
}

func Rename(src, dest gleam_P.String_t) gleam_P.Result_t[gleam_P.Nil_t, FileError_t] {
	return gleamResult(func() (gleam_P.Nil_t, error) {
		return gleam_P.Nil_c{}, os.Rename(path.Clean(string(src)), path.Clean(string(dest)))
	})
}

func SetPermissionsOctal(filepath gleam_P.String_t, permissions gleam_P.Int_t) gleam_P.Result_t[gleam_P.Nil_t, FileError_t] {
	return gleamResult(func() (gleam_P.Nil_t, error) {
		return gleam_P.Nil_c{}, os.Chmod(path.Clean(string(filepath)), os.FileMode(permissions))
	})
}

func CurrentDirectory() gleam_P.Result_t[gleam_P.String_t, FileError_t] {
	return gleamResult(func() (gleam_P.String_t, error) {
		res, err := os.Getwd()
		return gleam_P.String_t(res), err
	})
}

func toFileInfo(filepath gleam_P.String_t, stat os.FileInfo) (FileInfo_t, error) {
	sys, ok := stat.Sys().(*syscall.Stat_t)
	if !ok {
		return FileInfo_c{}, fmt.Errorf("file '%s': expected stat.Sys() to return value of type '*syscall.Stat_t' but got value of type '%T'", filepath, stat.Sys())
	}
	if sys == nil {
		return FileInfo_c{}, fmt.Errorf("file '%s': stat.Sys() returned nil", filepath)
	}
	return FileInfo_c{
		Size:         gleam_P.Int_t(stat.Size()),
		Mode:         gleam_P.Int_t(stat.Mode()),
		Nlinks:       gleam_P.Int_t(sys.Nlink),
		Inode:        gleam_P.Int_t(sys.Ino),
		UserId:       gleam_P.Int_t(sys.Uid),
		GroupId:      gleam_P.Int_t(sys.Gid),
		Dev:          gleam_P.Int_t(sys.Dev),
		AtimeSeconds: gleam_P.Int_t(sys.Atim.Sec),
		MtimeSeconds: gleam_P.Int_t(sys.Mtim.Sec),
		CtimeSeconds: gleam_P.Int_t(sys.Ctim.Sec),
	}, nil
}

func FileInfo(filepath gleam_P.String_t) gleam_P.Result_t[FileInfo_t, FileError_t] {
	return gleamResult(func() (FileInfo_t, error) {
		filepath := path.Clean(string(filepath))
		stat, err := os.Stat(filepath)
		if err != nil {
			return FileInfo_c{}, err
		}
		return toFileInfo(gleam_P.String_t(filepath), stat)
	})
}

func LinkInfo(filepath gleam_P.String_t) gleam_P.Result_t[FileInfo_t, FileError_t] {
	return gleamResult(func() (FileInfo_t, error) {
		filepath := path.Clean(string(filepath))
		stat, err := os.Lstat(filepath)
		if err != nil {
			return FileInfo_c{}, err
		}
		return toFileInfo(gleam_P.String_t(filepath), stat)
	})
}

func gleamResult[T gleam_P.Type[T]](op func() (T, error)) gleam_P.Result_t[T, FileError_t] {
	res, err := op()
	if err != nil {
		var errno syscall.Errno
		if errors.As(err, &errno) {
			return gleam_P.Error_c[T, FileError_t]{castError(errno)}
		}
		return gleam_P.Error_c[T, FileError_t]{Unknown_c{gleam_P.String_t(err.Error())}}
	}
	return gleam_P.Ok_c[T, FileError_t]{res}
}

func castError(errorCode syscall.Errno) FileError_t {
	switch errorCode {
	case syscall.EACCES:
		return Eacces_c{}
	case syscall.EAGAIN:
		return Eagain_c{}
	case syscall.EBADF:
		return Ebadf_c{}
	case syscall.EBADMSG:
		return Ebadmsg_c{}
	case syscall.EBUSY:
		return Ebusy_c{}
	// case syscall.EDEADLK:
	// 	return Edeadlk_c{}
	case syscall.EDEADLOCK:
		return Edeadlock_c{}
	case syscall.EDQUOT:
		return Edquot_c{}
	case syscall.EEXIST:
		return Eexist_c{}
	case syscall.EFAULT:
		return Efault_c{}
	case syscall.EFBIG:
		return Efbig_c{}
	// case syscall.EFTYPE:
	// 	return Eftype_c{}
	case syscall.EINTR:
		return Eintr_c{}
	case syscall.EINVAL:
		return Einval_c{}
	case syscall.EIO:
		return Eio_c{}
	case syscall.EISDIR:
		return Eisdir_c{}
	case syscall.ELOOP:
		return Eloop_c{}
	case syscall.EMFILE:
		return Emfile_c{}
	case syscall.EMLINK:
		return Emlink_c{}
	case syscall.EMULTIHOP:
		return Emultihop_c{}
	case syscall.ENAMETOOLONG:
		return Enametoolong_c{}
	case syscall.ENFILE:
		return Enfile_c{}
	case syscall.ENOBUFS:
		return Enobufs_c{}
	case syscall.ENODEV:
		return Enodev_c{}
	case syscall.ENOLCK:
		return Enolck_c{}
	case syscall.ENOLINK:
		return Enolink_c{}
	case syscall.ENOENT:
		return Enoent_c{}
	case syscall.ENOMEM:
		return Enomem_c{}
	case syscall.ENOSPC:
		return Enospc_c{}
	case syscall.ENOSR:
		return Enosr_c{}
	case syscall.ENOSTR:
		return Enostr_c{}
	case syscall.ENOSYS:
		return Enosys_c{}
	// case syscall.ENOBLK:
	// 	return Enotblk_c{}
	// case syscall.ENODIR:
	// 	return Enotdir_c{}
	// case syscall.ENOTSUP:
	// 	return Enotsup_c{}
	case syscall.ENXIO:
		return Enxio_c{}
	case syscall.EOPNOTSUPP:
		return Eopnotsupp_c{}
	case syscall.EOVERFLOW:
		return Eoverflow_c{}
	case syscall.EPERM:
		return Eperm_c{}
	case syscall.EPIPE:
		return Epipe_c{}
	case syscall.ERANGE:
		return Erange_c{}
	case syscall.EROFS:
		return Erofs_c{}
	case syscall.ESPIPE:
		return Espipe_c{}
	case syscall.ESRCH:
		return Esrch_c{}
	case syscall.ESTALE:
		return Estale_c{}
	case syscall.ETXTBSY:
		return Etxtbsy_c{}
	case syscall.EXDEV:
		return Exdev_c{}
	// case syscall.NOTUTF8:
	// 	return NotUtf8_c{}
	default:
		return Unknown_c{gleam_P.String_t(errorCode.Error())}
	}
}
