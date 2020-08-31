package tasks

import (
	"context"
	"errors"
	"fmt"
	"github.com/goftp/server"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	appExec "github.com/cloudradar-monitoring/tacoscript/exec"

	"github.com/cloudradar-monitoring/tacoscript/utils"

	filedriver "github.com/goftp/file-driver"

	"github.com/stretchr/testify/assert"
)

type fileManagedTestCase struct {
	Task            *FileManagedTask
	ExpectedResult  ExecutionResult
	RunnerMock      *appExec.SystemRunner
	Name            string
	FileShouldExist bool
	ContentToWrite  string
	FileExpectation *utils.FileExpectation
}

func TestFileManagedTaskExecution(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const ftpPort = 3021

	ftpUrl, err := startFTPServer(ctx, ftpPort)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	httpSrvUrl, httpSrv, err := startHTTPServer(false)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	defer httpSrv.Close()

	httpsSrvUrl, httpsSrv, err := startHTTPServer(true)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	defer httpsSrv.Close()

	filesToDelete := []string{
		"sourceFileHTTPS.txt",
		"sourceFileHTTP.txt",
		"sourceFileAtLocal.txt",
		"sourceFileFTP.txt",
	}

	err = ioutil.WriteFile("sourceFileAtLocal.txt", []byte("one two three"), 0644)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	err = ioutil.WriteFile("sourceFileHTTPS.txt", []byte("one two three"), 0644)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	httpsSrvUrl.Path = "/sourceFileHTTPS.txt"

	err = ioutil.WriteFile("sourceFileHTTP.txt", []byte("one two three"), 0644)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	httpSrvUrl.Path = "/sourceFileHTTP.txt"

	err = ioutil.WriteFile("sourceFileFTP.txt", []byte("one two three"), 0644)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	ftpUrl.Path = "sourceFileFTP.txt"

	testCases := []struct {
		Task            *FileManagedTask
		ExpectedResult  ExecutionResult
		RunnerMock      *appExec.SystemRunner
		Name            string
		FileShouldExist bool
		ContentToWrite  string
		FileExpectation *utils.FileExpectation
	}{
		{
			Name: "test creates field",
			Task: &FileManagedTask{
				Path:    "somepath",
				Name:    "some test command",
				Creates: []string{"some file 123", "some file 345"},
			},
			ExpectedResult: ExecutionResult{
				IsSkipped: true,
				Err:       nil,
			},
			FileShouldExist: true,
		},
		{
			Name: "test hash matches",
			Task: &FileManagedTask{
				Path:       "somepath",
				Name:       "someTempFile.txt",
				SourceHash: "md5=5e4fe0155703dde467f3ab234e6f966f",
			},
			ExpectedResult: ExecutionResult{
				IsSkipped: true,
				Err:       nil,
			},
			ContentToWrite: "one two three",
		},
		{
			Name: "test wrong hash format error",
			Task: &FileManagedTask{
				Path:       "somepath",
				Name:       "someTempFile.txt",
				SourceHash: "md4=5e4fe0155703dde467f3ab234e6f966f",
			},
			ExpectedResult: ExecutionResult{
				Err: errors.New("unknown hash algorithm 'md4' in 'md4=5e4fe0155703dde467f3ab234e6f966f'"),
			},
		},
		{
			Name: "local_source_copy_success",
			Task: &FileManagedTask{
				Path:       "local_source_copy_success_path",
				Name:       "targetFileAtLocal.txt",
				SourceHash: "md5=5e4fe0155703dde467f3ab234e6f966f",
				Source: utils.Location{
					IsURL:       false,
					LocalPath:   "sourceFileAtLocal.txt",
					RawLocation: "sourceFileAtLocal.txt",
				},
			},
			ExpectedResult: ExecutionResult{IsSkipped: false},
			FileExpectation: &utils.FileExpectation{
				FilePath:        "targetFileAtLocal.txt",
				ShouldExist:     true,
				ExpectedContent: "one two three",
			},
		},
		{
			Name: "http_source_copy_success",
			Task: &FileManagedTask{
				Path:       "http_source_copy_success_path",
				Name:       "targetFileFromHttp.txt",
				SourceHash: "md5=5e4fe0155703dde467f3ab234e6f966f",
				Source: utils.Location{
					IsURL:       true,
					Url:         httpSrvUrl,
					RawLocation: httpSrvUrl.String(),
				},
			},
			ExpectedResult: ExecutionResult{IsSkipped: false},
			FileExpectation: &utils.FileExpectation{
				FilePath:        "targetFileFromHttp.txt",
				ShouldExist:     true,
				ExpectedContent: "one two three",
			},
		},
		{
			Name: "https_source_copy_success",
			Task: &FileManagedTask{
				Path:       "https_source_copy_success_path",
				Name:       "targetFileFromHttps.txt",
				SourceHash: "md5=5e4fe0155703dde467f3ab234e6f966f",
				Source: utils.Location{
					IsURL:       true,
					Url:         httpsSrvUrl,
					RawLocation: httpsSrvUrl.String(),
				},
				SkipTlsCheck: true,
			},
			ExpectedResult: ExecutionResult{IsSkipped: false},
			FileExpectation: &utils.FileExpectation{
				FilePath:        "targetFileFromHttps.txt",
				ShouldExist:     true,
				ExpectedContent: "one two three",
			},
		},
		{
			Name: "ftp_source_copy_success",
			Task: &FileManagedTask{
				Path:       "ftp_source_copy_success_path",
				Name:       "targetFileFromFtp.txt",
				SourceHash: "md5=5e4fe0155703dde467f3ab234e6f966f",
				Source: utils.Location{
					IsURL:       true,
					Url:         ftpUrl,
					RawLocation: ftpUrl.String(),
				},
			},
			ExpectedResult: ExecutionResult{IsSkipped: false},
			FileExpectation: &utils.FileExpectation{
				FilePath:        "targetFileFromFtp.txt",
				ShouldExist:     true,
				ExpectedContent: "one two three",
			},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.Name, func(tt *testing.T) {
			assertTestCase(tt, tc)
			filesToDelete = append(filesToDelete, tc.Task.Name)
		})
	}

	err = deleteFiles(filesToDelete)
	if err != nil {
		log.Warn(err)
	}
}

func assertTestCase(t *testing.T, tc fileManagedTestCase) {
	if tc.ContentToWrite != "" {
		err := ioutil.WriteFile(tc.Task.Name, []byte(tc.ContentToWrite), 0644)
		assert.NoError(t, err)
	}
	runner := tc.RunnerMock
	if runner == nil {
		runner = &appExec.SystemRunner{SystemAPI: &appExec.SystemAPIMock{}}
	}
	fileManagedExecutor := &FileManagedTaskExecutor{
		Runner: runner,
		FsManager: &utils.FsManagerMock{
			ExistsToReturn: tc.FileShouldExist,
		},
	}

	res := fileManagedExecutor.Execute(context.Background(), tc.Task)
	assert.EqualValues(t, tc.ExpectedResult.Err, res.Err)
	assert.EqualValues(t, tc.ExpectedResult.IsSkipped, res.IsSkipped)
	assert.EqualValues(t, tc.ExpectedResult.StdOut, res.StdOut)
	assert.EqualValues(t, tc.ExpectedResult.StdErr, res.StdErr)

	if tc.ExpectedResult.Err != nil {
		return
	}

	if tc.ExpectedResult.IsSkipped {
		return
	}

	if tc.FileExpectation == nil {
		return
	}

	isExpectationMatched, nonMatchedReason, err := utils.AssertFileMatchesExpectation(tc.Task.Name, tc.FileExpectation)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	if !isExpectationMatched {
		assert.Fail(t, nonMatchedReason)
	}
}

func TestFileManagedTaskValidation(t *testing.T) {
	testCases := []struct {
		Task          FileManagedTask
		ExpectedError string
	}{
		{
			Task: FileManagedTask{
				Path: "somepath",
			},
			ExpectedError: fmt.Sprintf("empty required value at path 'somepath.%s'", NameField),
		},
		{
			Task: FileManagedTask{
				Name: "task1",
				Path: "somepath1",
				Source: utils.Location{
					IsURL: true,
					Url: &url.URL{
						Scheme: "http",
						Host:   "ya.ru",
					},
					RawLocation: "http://ya.ru",
				},
			},
			ExpectedError: fmt.Sprintf(`empty '%s' field at path 'somepath1.%s' for remote url source 'http://ya.ru'`, SourceHashField, SourceHashField),
		},
		{
			Task: FileManagedTask{
				Name: "task2",
				Path: "somepath2",
				Source: utils.Location{
					IsURL: true,
					Url: &url.URL{
						Scheme: "ftp",
						Host:   "ya.ru",
					},
					RawLocation: "ftp://ya.ru",
				},
			},
			ExpectedError: fmt.Sprintf(`empty '%s' field at path 'somepath2.%s' for remote url source 'ftp://ya.ru'`, SourceHashField, SourceHashField),
		},
		{
			Task: FileManagedTask{
				Name: "some p",
				Source: utils.Location{
					IsURL:     false,
					LocalPath: "/somepath",
				},
			},
		},
	}

	for _, testCase := range testCases {
		err := testCase.Task.Validate()
		if testCase.ExpectedError == "" {
			assert.NoError(t, err)
		} else {
			assert.EqualError(t, err, testCase.ExpectedError)
		}
	}
}

func deleteFiles(files []string) error {
	errs := &utils.Errors{
		Errs: []error{},
	}
	for _, file := range files {
		errs.Add(os.Remove(file))
	}

	return errs.ToError()
}

func startHTTPServer(isHttps bool) (u *url.URL, srv *httptest.Server, err error) {
	if isHttps {
		srv = httptest.NewTLSServer(http.FileServer(http.Dir(".")))
	} else {
		srv = httptest.NewServer(http.FileServer(http.Dir(".")))
	}

	u, err = url.Parse(srv.URL)

	return
}

func startFTPServer(ctx context.Context, port int) (*url.URL, error) {
	path, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	ftpHost := fmt.Sprintf("ftp://root:root@localhost:%d", port)
	ftpHostURL, err := url.Parse(ftpHost)
	if err != nil {
		return nil, err
	}

	factory := &filedriver.FileDriverFactory{
		RootPath: path,
		Perm:     server.NewSimplePerm("user", "group"),
	}

	opts := &server.ServerOpts{
		Factory:  factory,
		Port:     port,
		Hostname: "localhost",
		Auth:     &server.SimpleAuth{Name: "root", Password: "root"},
	}

	log.Printf("Starting ftp server on %v:%v", opts.Hostname, opts.Port)
	ftpSrvr := server.NewServer(opts)

	go func() {
		<-ctx.Done()
		err := ftpSrvr.Shutdown()
		if err != nil {
			log.Error(err)
		}
	}()

	errChan := make(chan error, 1)
	go func() {
		defer close(errChan)
		err := ftpSrvr.ListenAndServe()
		if err != nil {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return ftpHostURL, err
	case <-time.After(time.Millisecond * 300):
		return ftpHostURL, nil
	}
}