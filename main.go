package main

import (
	"os"
)

const (
	MAX_FILENAME_LENGTH = 256
	MAX_BYTES_TO_READ   = 2 * 1024 // 2KB buffer to read file
)

func main() {
	if len(os.Args) == 1 {
		usage()
	}

	for _, filename := range os.Args[1:] {
		fi, err := os.Lstat(filename)
		if err != nil {
			print(filename + ": " + err.Error())
			continue
		}

		if len(filename) > MAX_FILENAME_LENGTH {
			print("File name too long.")
			continue
		}

		print(filename + ": ")

		if fi.Mode()&os.ModeSymlink != 0 {
			reallink, _ := os.Readlink(filename)
			print("symbolic link to " + reallink)
		} else if fi.Mode()&os.ModeDir != 0 {
			print("directory")
		} else if fi.Mode()&os.ModeSocket != 0 {
			print("socket")
		} else if fi.Mode()&os.ModeCharDevice != 0 {
			print("character special device")
		} else if fi.Mode()&os.ModeDevice != 0 {
			print("device file")
		} else if fi.Mode()&os.ModeNamedPipe != 0 {
			print("fifo")
		} else if fi.Mode()&os.ModeIrregular == 0 {
			regularFile(filename)
		} else {
			print("unknown")
		}
		println()
	}
}

func checkerr(e error) {
	if e != nil {
		print(e.Error())
		os.Exit(1)
	}
}

func usage() {
	println("Usage: figo FILE_NAME")
	os.Exit(0)
}

func regularFile(filename string) {

	/*---------------Read file------------------------*/
	file, _ := os.OpenFile(filename, os.O_RDONLY, 0666)
	// checkerr(err)
	defer file.Close()

	var contentByte = make([]byte, MAX_BYTES_TO_READ) //

	numByte, _ := file.Read(contentByte)
	// if err != nil && err != io.EOF {
	// 	checkerr(err)
	// }
	contentByte = contentByte[:numByte]

	contentStr := string(contentByte) // File content in string
	lenb := len(contentByte)
	/*---------------Read file end------------------------*/
	magic := -1
	if lenb > 112 {
		magic = int(peekLe(contentStr[60:], 4))
	}

	if lenb >= 45 && HasPrefix(contentStr, "\177ELF") {
		print("Elf file ")
		doElf(contentByte)
	} else if lenb >= 8 && HasPrefix(contentStr, "!<arch>\n") {
		print("ar archive")
	} else if lenb > 28 && HasPrefix(contentStr, "\x89PNG\x0d\x0a\x1a\x0a") {
		print("PNG image data")
	} else if lenb > 16 &&
		(HasPrefix(contentStr, "GIF87a") || HasPrefix(contentStr, "GIF89a")) {
		print("GIF image data")
	} else if lenb > 32 && HasPrefix(contentStr, "\xff\xd8") {
		print("JPEG image data")
	} else if lenb > 8 && HasPrefix(contentStr, "\xca\xfe\xba\xbe") {
		print("JAVA class file")
	} else if lenb > 8 && HasPrefix(contentStr, "dex\n") {
		print("Android dex file")
	} else if lenb > 500 && contentStr[257:262] == "ustar" {
		print("Posix tar archive")
	} else if lenb > 5 && HasPrefix(contentStr, "PK\x03\x04") {
		print("Zip archive data")
	} else if lenb > 4 && HasPrefix(contentStr, "BZh") {
		print("bzip2 compressed data")
	} else if lenb > 10 && HasPrefix(contentStr, "\x1f\x8b") {
		print("gzip compressed data")
	} else if lenb > 32 && contentStr[1:4] == "\xfa\xed\xfe" {
		print("Mach-O")
	} else if lenb > 36 && contentStr[0:6] == "OggS\x00\x02" {
		print("Ogg data")
	} else if lenb > 32 && contentStr[0:3] == "RIF" &&
		string(contentByte[8:16]) == "WAVEfmt " {
		print("WAV audio")
	} else if lenb > 12 && HasPrefix(contentStr, "\x00\x01\x00\x00") {
		print("TrueType font")
	} else if lenb > 12 && HasPrefix(contentStr, "ttcf\x00") {
		print("TrueType font collection")
	} else if lenb > 4 && HasPrefix(contentStr, "BC\xc0\xde") {
		print("LLVM IR bitcode")
	} else if HasPrefix(contentStr, "-----BEGIN CERTIFICATE-----") {
		print("PEM certificate")
	} else if magic != -1 && HasPrefix(contentStr, "MZ") && magic < lenb-4 &&
		contentStr[magic:magic+4] == "\x50\x45\x00\x00" {

		print("MS executable")
		if int(peekLe(contentStr[magic+22:], 2)&0x2000) != 0 {
			print("(DLL)")
		}
		print(" ")
		if peekLe(contentStr[magic+20:], 2) > 70 {
			types := []string{"", "native", "GUI", "console", "OS/2", "driver", "CE",
				"EFI", "EFI boot", "EFI runtime", "EFI ROM", "XBOX", "", "boot"}
			tp := int(peekLe(contentStr[magic+92:], 2))
			if tp > 0 && tp < len(types) {
				print(types[tp])
			}
		}
	} else if lenb > 50 && HasPrefix(contentStr, "BM") &&
		contentStr[6:10] == "\x00\x00\x00\x00" {
		print("BMP image")
	}
}

func doElf(contentByte []byte) {
	contentStr := string(contentByte)
	contentChar := []byte(contentStr)
	bits := int(contentChar[4])
	endian := contentChar[5]

	var elfint func(str string, size int) int64

	if endian == 2 {
		elfint = peekBe
	} else {
		elfint = peekLe
	}

	exei := elfint(contentStr[16:], 2)
	if exei == 1 {
		print("relocatable")
	} else if exei == 2 {
		print("executable")
	} else if exei == 3 {
		print("shared object")
	} else if exei == 4 {
		print("core dump")
	} else {
		print("bad type")
	}

	print(", ")

	if bits == 1 {
		print("32bit ")
	} else if bits == 2 {
		print("64bit ")
	}

	if endian == 1 {
		print("LSB ")
	} else if endian == 2 {
		print("MSB ")
	} else {
		print("bad endian ")
	}

	/* You can have a full list from here https://golang.org/src/debug/elf/elf.go */
	archType := map[string]int{
		"alpha": 0x9026, "arc": 93, "arcv2": 195, "arm": 40, "arm64": 183,
		"avr32": 0x18ad, "bpf": 247, "blackfin": 106, "c6x": 140, "cell": 23,
		"cris": 76, "frv": 0x5441, "h8300": 46, "hexagon": 164, "ia64": 50,
		"m32r88": 88, "m32r": 0x9041, "m68k": 4, "metag": 174, "microblaze": 189,
		"microblaze-old": 0xbaab, "mips": 8, "mips-old": 10, "mn10300": 89,
		"mn10300-old": 0xbeef, "nios2": 113, "openrisc": 92, "openrisc-old": 0x8472,
		"parisc": 15, "ppc": 20, "ppc64": 21, "s390": 22, "s390-old": 0xa390,
		"score": 135, "sh": 42, "sparc": 2, "sparc8+": 18, "sparc9": 43, "tile": 188,
		"tilegx": 191, "386": 3, "486": 6, "x86-64": 62, "xtensa": 94, "xtensa-old": 0xabc7,
	}

	archj := elfint(contentStr[18:], 2)
	for key, val := range archType {
		if val == int(archj) {
			print(key)
			break
		}
	}

	bits--

	phentsize := elfint(contentStr[42+12*bits:], 2)
	phnum := elfint(contentStr[44+12*bits:], 2)
	phoff := elfint(contentStr[28+4*bits:], 4+4*bits)
	// shsize 		:= elfint(contentStr[46+12*bits:], 2)
	// shnum 		:= elfint(contentStr[48+12*bits:], 2)
	// shoff 		:= elfint(contentStr[32+8*bits:], 4+4*bits)
	dynamic := false

	for i := 0; i < int(phnum); i++ {
		phdr := contentStr[int(phoff)+i*int(phentsize):]
		// char *phdr = map+phoff+i*phentsize;
		p_type := elfint(phdr, 4)

		dynamic = (p_type == 2) || dynamic /*PT_DYNAMIC*/
		if p_type != 3 /*PT_INTERP*/ && p_type != 4 /*PT_NOTE*/ {
			continue
		}

		// j = bits+1
		// p_offset := elfint(phdr[4*j:], 4*j)
		// p_filesz := elfint(phdr[16*j:], 4*j)

		if p_type == 3 /*PT_INTERP*/ {
			print(", dynamically linked")
			//   print(p_filesz)
			//   print(contentStr[p_offset*2:])
		}
	}

	if !dynamic {
		print(", statically linked")
	}
}

func HasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[0:len(prefix)] == prefix
}

func peekLe(str string, size int) int64 {
	ret := int64(0)
	c := []byte(str)

	for i := 0; i < size; i++ {
		ret = ret | int64(c[i])<<uint8(i*8)
	}
	return ret
}

func peekBe(str string, size int) int64 {
	ret := int64(0)
	c := []byte(str)

	for i := 0; i < size; i++ {
		ret = (ret << 8) | (int64(c[i]) & 0xff)
	}
	return ret
}
