# utils
--
    import "cloud/cloud/utils"


## Usage

```go
var (
	LogLevels = []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR", "CRITICAL"}
)
```

#### func  AvailableDisk

```go
func AvailableDisk(path string) uint64
```

#### func  DirSize

```go
func DirSize(path string) (uint64, error)
```
https://stackoverflow.com/questions/32482673/how-to-get-directory-total-size

#### func  GetLogger

```go
func GetLogger() *log.Logger
```
GetLogger returns a global logger variable, or creates a new default logger.

#### func  GetTestDirs

```go
func GetTestDirs(prefix string, n int) ([]string, error)
```
GetTestDirs creates n temporary directories using the given prefix. It returns
the path of all the created directories. It is the caller's responsibility to
remove the temporary directories. The caller may call the GetTestDirsCleanup
with defer for appropriate clean up.

#### func  GetTestDirsCleanup

```go
func GetTestDirsCleanup(dirs []string)
```
GetTestDirsCleanup performs clean up for GetTestDirs.

#### func  GetTestFile

```go
func GetTestFile(prefix string, contents []byte) (*os.File, error)
```
GetTestFile returns a temporary file with the given byte contents and a filename
prefix. It is the caller's responsibility to remove the temporary file. The
caller may call the GetTestFileCleanup with defer for appropriate clean up.

#### func  GetTestFileCleanup

```go
func GetTestFileCleanup(file *os.File)
```
GetTestFileCleanup performs clean up for GetTestFile.

#### func  HashFile

```go
func HashFile(buffer []byte) string
```
HashFile computes the hash of a file's bytes using a suitable hash function. The
function returns the hash as a hex-encoded string.

#### func  MaxInt

```go
func MaxInt(nums []int) (int, int, error)
```
Max returns the maximum element and its index in an integer slice. An error is
raised if the length of the slice is 0.

#### func  NewLoggerFromEnv

```go
func NewLoggerFromEnv() *log.Logger
```
NewLoggerFromEnv creates a new logger from environment variables.

#### func  NewLoggerFromLevel

```go
func NewLoggerFromLevel(level string) *log.Logger
```
NewLoggerFromLevel creates a new logger using a certain log level, using the
default writer.

#### func  NewLoggerFromWriter

```go
func NewLoggerFromWriter(writer io.Writer) *log.Logger
```
NewLoggerFromWriter creates a new logger from a writer, using the default level.

#### func  NewLoggerFromWriterLevel

```go
func NewLoggerFromWriterLevel(writer io.Writer, level string) *log.Logger
```
NewLoggerFromWriterLevel creates a new global logger using the given writer and
level. Level should be one of the levels found in the LogLevels variable. The
old logger variable is overriden.

#### func  RemoveDirs

```go
func RemoveDirs(dirs []string)
```
RemoveDirs removes all the directories in the list.
