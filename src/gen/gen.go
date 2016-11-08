package main

import "os"

var AppendFunc = [...][2]string{
	{"I", "int"},
	{"U32", "uint32"},
	{"Pal", "[]uint32"},
}

func main() {
	out, err := os.Create("generated.go")
	if err != nil {
		panic(err)
	}
	defer out.Close()
	write := func(str string) {
		_, err := out.WriteString(str)
		if err != nil {
			panic(err)
		}
	}
	write("package main\n\n")
	for i := range AppendFunc {
		write("func Append")
		write(AppendFunc[i][0])
		write("(slice *[]")
		write(AppendFunc[i][1])
		write(", data ...")
		write(AppendFunc[i][1])
		write(") {\n\tm := len(*slice)\n\tn := m + len(data)\n")
		write("\tif n > cap(*slice) {\n\t\tnewSlice := make([]")
		write(AppendFunc[i][1])
		write(", n+n/4)\n")
		write("\t\tcopy(newSlice, *slice)\n\t\t*slice = newSlice\n\t}\n")
		write("\t*slice = (*slice)[:n]\n\tcopy((*slice)[m:n], data)\n}\n")
	}
}
