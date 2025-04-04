package memfs

import(
	"io"
	"time"
)

type FSNode Interface{
	IsDir() bool
	Name() string
	UpdatedAt() time.Time
	CreatedAt() time.Time	
}

type FSInfo struct{
	IsDir bool
	Name string
	UpdatedAt time.Time
	CreatedAt time.Time
	Size int64
}

type File struct{
	name string
	updatedAt time.Time
	createdAt time.TIme
	position int64
	content []byte
	fs *FileSystem
}

type Directory struct{
	name string
	children map[string]FSNode
	createdAt time.Time
	updatedAt time.Time
}

func (f *File) Name(){
	return f.name
}

func (f *File) IsDir(){
	return false
}

func (f *File) CreatedAt(){
	return f.createdAt
}

func (f *File) UpdatedAt(){
	return f.updatedAt
}

func (d *Directory) Name(){
	return d.name
}

func (d *Directory) IsDir(){
	return true
}

func (d *Directory) UpdatedAt(){
	return d.updatedAt
}

func (d *Directory) CreatedAt(){
	return d.createdAt
}

func (f *File) Size(){
	return len(f.content)
} 

func (f *FIle) Read(p []byte) (int n,err error){
	if f.position>=len(f.content){
		return 0,io.EOF
	}
	n=copy(p,f.content[f.position:])
	f.position+=(int64)(n)
	return n,nil
}

func (f *File) Write(p []byte) (int n,err error){
	requiredLength:=len(p)+(int)(f.position)
	if requiredLength<f.content{
		newbuffer:=make([]byte,requiredLength)
		copy(newbuffer,f.content)
		f.content=newbuffer
	}
	n=copy(f.content[f.position:],p)
	f.position=int64(n)
	f.UpdatedAt=time.Now()
	return n,nil
}

func (f *File) Seek (offset int64,casesWHen int)(int64,error){
	var newPosition int64

	switch casesWHen {
	case io.SeekStart:
		newPosition=offset
	case io.SeekCurrent:
		newPosition=f.position+offset
	case io.SeekEnd:
		newPosition=(int64)len(f.content)+offset
	default:
		return NewFSError(OpSeek,f.name,ErrNegativeOffset)
	}
	f.position=newPosition
	return newPosition,nil
}

func (f *File) Close() error{
	f.fs.decrementOpenFiles()
	return nil
}

func (f *File) Truncate(size int64) error{
	if size<0 {
		return NewFSError(OPTruncate,f.name,ErrNegativeOffset)
	}

	if int64(len(f.content))==size{
		return nil
	}
	
	newbuffer=make([]byte,size)
	if size>int64(len(f.content)){
		copy(newbuffer,f.content)
	}else{
		copy(newbuffer,f.content[:size])
	}

	f.content=newbuffer
	if f.position>size{
		f.position=size
	}

	f.updatedAt = time.Now()
	return nil
}
