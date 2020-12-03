package scenezip

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

var id int

// ExtractFromZipToTxt 将src目录中的zip格式的场景文件转成txt格式的文件,并存入dst目录中
func ExtractFromZipToTxt(srcDir, dstDir string) error {
	//! 家目录
	home := filepath.Dir(srcDir)

	//! srcDir目录中的:场景.zip 文件 --> tempScene目录中的:id_scene_xxxx.zip文件
	// 创建临时目录
	tempScene, err := ioutil.TempDir(home, "tempScene_")
	if err != nil {
		return errors.Wrapf(err, "scenezip: create tempdir %s failed", tempScene)
	}
	// 解压到临时目录
	if err := ExtractAll(srcDir, tempScene); err != nil {
		return err
	}
	// 最后删除临时目录
	defer os.RemoveAll(tempScene)

	//! tempScene目录中zip文件 --> tempD目录中的目录文件:场景名称_xxx
	// 创建临时目录
	tempD, err := ioutil.TempDir(home, "tempD_")
	if err != nil {
		return errors.Wrapf(err, "scenezip: create tempdir %s failed", tempD)
	}
	// 解压到临时目录
	if err := ExtractAll(tempScene, tempD); err != nil {
		return err
	}
	// 最后删除临时目录
	defer os.RemoveAll(tempD)

	//! 递归遍历tempD目录中的zip文件 --> dstDir目录中的文件:场景名称.txt
	// 解压到dstDir
	//* 如果dstDir存在,删除
	if err := os.RemoveAll(dstDir); err != nil {
		return errors.Wrapf(err, "scenezip: del dir %s failed", dstDir)
	}
	if err := os.MkdirAll(dstDir, os.ModePerm); err != nil {
		return errors.Wrapf(err, "scenezip: create dir %s failed", dstDir)
	}
	if err := ExtractAll(tempD, dstDir); err != nil {
		return err
	}

	// END
	return nil
}

// ReadFiles 递归读取目录中文件,返回文件绝对路径列表
func ReadFiles(dir string) ([]string, error) {
	var files = make([]string, 0)
	fis, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, errors.Wrapf(err, "scenezip: read dir %s failed", dir)
	}

	for _, fi := range fis {
		if fi.IsDir() {
			subdir := filepath.Join(dir, fi.Name())
			subfiles, err := ReadFiles(subdir)
			if err != nil {
				return nil, err
			}
			files = append(files, subfiles...)
			continue
		}
		filepath := filepath.Join(dir, fi.Name())
		files = append(files, filepath)
	}

	// END
	return files, nil
}

// ReadZipFiles 递归读取目录中文件,返回zip文件绝对路径列表
func ReadZipFiles(dir string) ([]string, error) {
	var zipfiles = make([]string, 0)

	files, err := ReadFiles(dir)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if strings.HasSuffix(file, "zip") {
			zipfiles = append(zipfiles, file)
		}
	}
	return zipfiles, nil
}

// Unzip 解压zip文件,至输出目录中
func Unzip(zipFile string, dstDir string) error {
	if !strings.HasSuffix(zipFile, "zip") {
		return errors.New("scenezip: not zip file")
	}
	zipReader, err := zip.OpenReader(zipFile)
	if err != nil {
		return errors.Wrapf(err, "scenezip: open zipfile %s failed", zipFile)
	}

	defer zipReader.Close()

	for _, f := range zipReader.File {
		fname := f.Name
		if strings.HasPrefix(fname, "tr_conv") {
			//! tr_conv,这个是场景的txt文件,需要将文件名称改为:场景名字.txt
			b := filepath.Base(zipFile)
			fname = strings.ReplaceAll(b, "zip", "txt")
		} else if strings.HasPrefix(fname, "tr_entity") || strings.HasPrefix(fname, "tr_config") || strings.HasSuffix(fname, ".dat") {
			//! 不需要的文件,就不需要提取了.
			continue
		} else if strings.HasPrefix(fname, "scene_") {
			//! scene_是中间的zip文件,需要对其进行区分,唯一化.
			fname = fmt.Sprintf("%d_%s", id, fname)
			id++
		}

		fpath := filepath.Join(dstDir, fname)
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fpath, os.ModePerm); err != nil {
				return errors.Wrapf(err, "scenezip: decompress %s failed", fpath)
			}
		} else {
			//? 解压缩至的目录若不存在
			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				errors.Wrapf(err, "scenezip: check decompress dir failed", filepath.Dir(fpath))
			}

			//! 打开zip压缩包中的文件
			inFile, err := f.Open()
			if err != nil {
				return errors.Wrap(err, "scenezip: open the file in zip error")
			}
			defer inFile.Close()

			//! 创建输出文件

			outFile, err := os.Create(fpath)
			if err != nil {
				return errors.Wrap(err, "scenezip: create dst zipfile failed")
			}
			defer outFile.Close()

			_, err = io.Copy(outFile, inFile)
			if err != nil {
				return errors.Wrap(err, "scenezip: copy file failed")
			}

		}
	}
	// END
	return nil
}

// ExtractAll 解压缩压缩目录中的zip到目标目录
func ExtractAll(srcDir, dstDir string) error {
	zipfiles, err := ReadZipFiles(srcDir)
	if err != nil {
		return err
	}

	for _, zipfile := range zipfiles {
		Unzip(zipfile, dstDir)
	}
	return nil
}
