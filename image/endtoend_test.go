package stegimg

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
	"testing"
)

const TestFolderName = "test"
const TestImgName = "testimg.png"
const OutputImgName = "output.png"
const TestFilePrefix = "testfile_"
const NumOfFilesToGenerate = 5
const ImageSize = 5000

func TestEncodeDecode(t *testing.T) {
	chdirToTestFileDir(TestFolderName)

	for LSBsToUse := byte(1); LSBsToUse <= 1; LSBsToUse++ {
		generateImageFile(TestImgName, ImageSize, ImageSize)

		testFiles := generateTestFiles(NumOfFilesToGenerate, ((ImageSize*ImageSize)*int(LSBsToUse*3)-(len(TestFilePrefix)+4*64*NumOfFilesToGenerate+NumOfFilesToGenerate*64))/(NumOfFilesToGenerate*8))
		originalHashes := calculateFileHashes(testFiles)
		err := Encode(TestImgName, OutputImgName, testFiles, Config{LSBsToUse: LSBsToUse})
		if err != nil {
			t.Fatal("Error encoding image")
		}

		deleteFiles(testFiles)

		DecodeImg(OutputImgName)
		decodedHashes := calculateFileHashes(testFiles)

		for i := 0; i < len(testFiles); i++ {
			if originalHashes[i] != decodedHashes[i] {
				t.Errorf("Hash for file %d is not the same after decoding - %s != %s | using %d LSBs", i, originalHashes[i], decodedHashes[i], LSBsToUse)
			}
		}
	}

	os.Chdir("..")
	os.RemoveAll(TestFolderName)
}

func BenchmarkFullEncodeSpeed(b *testing.B) {
	chdirToTestFileDir(TestFolderName)
	generateImageFile(TestImgName, ImageSize, ImageSize)
	for LSBsToUse := byte(1); LSBsToUse <= 8; LSBsToUse++ {
		for _, numOfBytesToEncode := range []int{100000, 1000000, 10000000} {
			filesToHide := generateTestFiles(NumOfFilesToGenerate, numOfBytesToEncode/NumOfFilesToGenerate)
			b.Run(fmt.Sprintf("MBs=%f/LSBsToUse=%d", float64(numOfBytesToEncode)/1000000.0, LSBsToUse), func(b *testing.B) {
				b.SetBytes(int64(numOfBytesToEncode))
				for i := 0; i < b.N; i++ {
					_ = Encode(TestImgName, OutputImgName, filesToHide, Config{LSBsToUse: LSBsToUse})
				}
			})
		}
	}
	os.Chdir("..")
	os.RemoveAll(TestFolderName)
}

func BenchmarkFullDecodeSpeed(b *testing.B) {
	chdirToTestFileDir(TestFolderName)
	generateImageFile(TestImgName, ImageSize, ImageSize)
	for LSBsToUse := byte(1); LSBsToUse <= 8; LSBsToUse++ {
		for _, numOfBytesToEncode := range []int{100000, 1000000, 10000000} {
			filesToHide := generateTestFiles(NumOfFilesToGenerate, numOfBytesToEncode/NumOfFilesToGenerate)
			_ = Encode(TestImgName, OutputImgName, filesToHide, Config{LSBsToUse: LSBsToUse})
			b.Run(fmt.Sprintf("MBs=%f/LSBsToUse=%d", float64(numOfBytesToEncode)/1000000.0, LSBsToUse), func(b *testing.B) {
				b.SetBytes(int64(numOfBytesToEncode))
				for i := 0; i < b.N; i++ {
					DecodeImg(OutputImgName)
				}
			})
		}
	}
	os.Chdir("..")
	os.RemoveAll(TestFolderName)
}

func deleteFiles(filesToDelete []string) {
	for _, file := range filesToDelete {
		err := os.Remove(file)
		if err != nil {
			panic(err)
		}
	}
}

func calculateFileHashes(fileNames []string) []string {
	hashes := make([]string, len(fileNames), len(fileNames))

	for idx, fileName := range fileNames {
		f, err := os.Open(fileName)
		if err != nil {
			log.Fatal(err)
		}

		h := sha256.New()
		if _, err := io.Copy(h, f); err != nil {
			log.Fatal(err)
		}
		hashes[idx] = fmt.Sprintf("%x", h.Sum(nil))
		f.Close()
	}

	return hashes
}
