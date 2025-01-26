package http_test

import (
	"crypto/rand"
	"fmt"
	"io"
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/ethanrous/weblens/internal/env"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/tests"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models"
	"github.com/ethanrous/weblens/models/rest"
	"github.com/ethanrous/weblens/service/proxy"
	"github.com/stretchr/testify/require"
)

func TestFiles(t *testing.T) {
	t.Parallel()

	coreServices, err := tests.NewWeblensTestInstance(t.Name(), env.Config{
		Role:     string(models.CoreServerRole),
		LogLevel: log.DEBUG,
	})

	require.NoError(t, err)

	keys, err := coreServices.AccessService.GetKeysByUser(coreServices.UserService.Get("test-username"))
	if err != nil {
		log.ErrTrace(err)
		t.FailNow()
	}

	coreInstance := coreServices.InstanceService.GetLocal()
	coreAddress := env.GetProxyAddress(coreServices.Cnf)
	coreApiKey := keys[0].Key

	localCoreInstance := models.NewInstance(coreInstance.Id, coreInstance.Name, coreApiKey, models.CoreServerRole, false, coreAddress, "")
	owner := coreServices.UserService.Get("test-username")
	if owner == nil {
		t.Fatalf("No owner")
	}

	err = simpleCreate(localCoreInstance, owner)
	if err != nil {
		log.ErrTrace(err)
		t.FailNow()
	}

	err = uploadFile(localCoreInstance, owner)
	if err != nil {
		log.ErrTrace(err)
		t.FailNow()
	}

	err = moveFiles(localCoreInstance, owner)
	if err != nil {
		log.ErrTrace(err)
		t.FailNow()
	}

	ownerHomeInfoRequest := proxy.NewCoreRequest(localCoreInstance, "GET", "/folder/"+owner.HomeId)
	ownerHomeInfo, err := proxy.CallHomeStruct[rest.FolderInfoResponse](ownerHomeInfoRequest)
	if err != nil {
		log.ErrTrace(err)
		t.FailNow()
	}

	if len(ownerHomeInfo.Children) != 1 || ownerHomeInfo.Children[0].Id != owner.TrashId {
		t.Fatalf("Owner home should be empty: %v", ownerHomeInfo.Children)
	}
}

func simpleCreate(core *models.Instance, owner *models.User) error {
	createFolderRequest := proxy.NewCoreRequest(core, "POST", "/folder").WithBody(rest.CreateFolderBody{
		ParentFolderId: owner.HomeId,
		NewFolderName:  "test-folder",
	})
	folderInfo, err := proxy.CallHomeStruct[rest.FileInfo](createFolderRequest)
	if err != nil {
		return err
	}

	if folderInfo.ParentId != owner.HomeId {
		return werror.Errorf("Parent folder mismatch %s != %s", folderInfo.ParentId, owner.HomeId)
	}

	getFileRequest := proxy.NewCoreRequest(core, "GET", "/files/"+folderInfo.Id)
	getFolderInfo, err := proxy.CallHomeStruct[rest.FileInfo](getFileRequest)
	if err != nil {
		return err
	}

	if getFolderInfo.Filename != "test-folder" {
		return werror.Errorf("Folder name mismatch %s != %s", getFolderInfo.Filename, "test-folder")
	}

	_, err = proxy.NewCoreRequest(core, "DELETE", "/files").WithQuery("ignore_trash", "true").WithBody(rest.FilesListParams{FileIds: []string{getFolderInfo.Id}}).Call()
	if err != nil {
		return err
	}

	return nil
}

func uploadFile(core *models.Instance, owner *models.User) error {
	randomBytes, fileSize, err := makeRandomFile()
	if err != nil {
		return err
	}

	createUploadRequest := proxy.NewCoreRequest(core, "POST", "/upload").WithBody(rest.NewUploadParams{
		RootFolderId: owner.HomeId,
		ChunkSize:    fileSize,
	})
	uploadInfo, err := proxy.CallHomeStruct[rest.NewUploadInfo](createUploadRequest)
	if err != nil {
		return err
	}

	newFileRequest := proxy.NewCoreRequest(core, "POST", "/upload/"+uploadInfo.UploadId).WithBody(rest.NewFilesParams{
		NewFiles: []rest.NewFileParams{
			{
				NewFileName:    "test-file.txt",
				ParentFolderId: owner.HomeId,
				FileSize:       fileSize,
			},
		},
	})
	newFilesInfo, err := proxy.CallHomeStruct[rest.NewFilesInfo](newFileRequest)
	if err != nil {
		return err
	}

	_, err = proxy.NewCoreRequest(core, "PUT", "/upload/"+uploadInfo.UploadId+"/file/"+newFilesInfo.FileIds[0]).WithHeader("Content-Range", fmt.Sprintf("0-%d/%d", fileSize, fileSize)).WithBodyBytes(randomBytes).Call()
	if err != nil {
		return err
	}

	getFileRequest := proxy.NewCoreRequest(core, "GET", "/files/"+newFilesInfo.FileIds[0])
	getFileInfo, err := proxy.CallHomeStruct[rest.FileInfo](getFileRequest)
	if err != nil {
		return err
	}

	if getFileInfo.Filename != "test-file.txt" {
		return werror.Errorf("File name mismatch %s != %s", getFileInfo.Filename, "test-file.txt")
	}
	if getFileInfo.ParentId != owner.HomeId {
		return werror.Errorf("Parent folder mismatch %s != %s", getFileInfo.ParentId, owner.HomeId)
	}

	returnedSize := getFileInfo.Size
	timeout := 10
	for returnedSize != fileSize {
		log.Debug.Println("Waiting for file upload to complete")
		time.Sleep(100 * time.Millisecond)
		timeout--
		if timeout == 0 {
			return werror.Errorf("Did not receive expected file size %d != %d", returnedSize, fileSize)
		}
		getFileRequest := proxy.NewCoreRequest(core, "GET", "/files/"+newFilesInfo.FileIds[0])
		getFileInfo, err := proxy.CallHomeStruct[rest.FileInfo](getFileRequest)
		if err != nil {
			return err
		}

		returnedSize = getFileInfo.Size
	}

	downloadUrl := fmt.Sprintf("/files/%s/download", newFilesInfo.FileIds[0])
	downloadResponse, err := proxy.NewCoreRequest(core, "GET", downloadUrl).Call()
	if err != nil {
		return err
	}

	bodyBytes, err := io.ReadAll(downloadResponse.Body)
	if err != nil {
		return err
	}

	if len(bodyBytes) != int(fileSize) {
		return werror.Errorf("Downloaded file size mismatch %d != %d", len(bodyBytes), fileSize)
	}
	if string(bodyBytes) != string(randomBytes) {
		return werror.Errorf("Downloaded file content mismatch")
	}

	_, err = proxy.NewCoreRequest(core, "DELETE", "/files").WithQuery("ignore_trash", "true").WithBody(rest.FilesListParams{FileIds: []string{newFilesInfo.FileIds[0]}}).Call()
	if err != nil {
		return err
	}

	return nil
}

func moveFiles(core *models.Instance, owner *models.User) error {
	folder1, err := makeFolder("top-folder-1", owner.HomeId, core)
	if err != nil {
		return err
	}

	folder2, err := makeFolder("top-folder-2", owner.HomeId, core)
	if err != nil {
		return err
	}

	subFolder, err := makeFolder("sub-folder", folder1.Id, core)
	if err != nil {
		return err
	}

	bytes, fileSize, err := makeRandomFile()
	if err != nil {
		return err
	}

	createUploadRequest := proxy.NewCoreRequest(core, "POST", "/upload").WithBody(rest.NewUploadParams{
		RootFolderId: subFolder.Id,
		ChunkSize:    fileSize,
	})

	uploadInfo, err := proxy.CallHomeStruct[rest.NewUploadInfo](createUploadRequest)
	if err != nil {
		return err
	}

	newFileRequest := proxy.NewCoreRequest(core, "POST", "/upload/"+uploadInfo.UploadId).WithBody(rest.NewFilesParams{
		NewFiles: []rest.NewFileParams{
			{
				NewFileName:    "test-file.txt",
				ParentFolderId: subFolder.Id,
				FileSize:       fileSize,
			},
		},
	})

	newFilesInfo, err := proxy.CallHomeStruct[rest.NewFilesInfo](newFileRequest)
	if err != nil {
		return err
	}

	_, err = proxy.NewCoreRequest(core, "PUT", "/upload/"+uploadInfo.UploadId+"/file/"+newFilesInfo.FileIds[0]).WithHeader("Content-Range", fmt.Sprintf("0-%d/%d", fileSize, fileSize)).WithBodyBytes(bytes).Call()
	if err != nil {
		return err
	}

	// Wait for upload to complete
	_, err = proxy.NewCoreRequest(core, "GET", "/upload/"+uploadInfo.UploadId).Call()
	if err != nil {
		return err
	}

	_, err = proxy.NewCoreRequest(core, "PATCH", "/files/").WithBody(rest.MoveFilesParams{
		NewParentId: folder2.Id,
		Files:       []string{subFolder.Id},
	}).Call()
	if err != nil {
		return err
	}

	getFolder1Request := proxy.NewCoreRequest(core, "GET", "/folder/"+folder1.Id)
	folder1Info, err := proxy.CallHomeStruct[rest.FolderInfoResponse](getFolder1Request)
	if err != nil {
		return err
	}

	if len(folder1Info.Children) != 0 {
		return werror.Errorf("Folder 1 should be empty")
	}

	getFolder2Request := proxy.NewCoreRequest(core, "GET", "/folder/"+folder2.Id)
	folder2Info, err := proxy.CallHomeStruct[rest.FolderInfoResponse](getFolder2Request)
	if err != nil {
		return err
	}

	if len(folder2Info.Children) != 1 {
		return werror.Errorf("Folder 2 should have 1 child")
	}

	if folder2Info.Children[0].Id != subFolder.Id {
		return werror.Errorf("Folder 2 should have sub-folder as child")
	}

	_, err = proxy.NewCoreRequest(core, "DELETE", "/files").WithQuery("ignore_trash", "true").WithBody(rest.FilesListParams{FileIds: []string{folder1.Id, folder2.Id}}).Call()
	if err != nil {
		return err
	}

	return nil
}

func makeFolder(name string, parentFolderId string, core *models.Instance) (rest.FileInfo, error) {
	createFolderRequest := proxy.NewCoreRequest(core, "POST", "/folder").WithBody(rest.CreateFolderBody{
		ParentFolderId: parentFolderId,
		NewFolderName:  name,
	})
	folderInfo, err := proxy.CallHomeStruct[rest.FileInfo](createFolderRequest)
	if err != nil {
		return rest.FileInfo{}, err
	}

	return folderInfo, nil
}

func makeRandomFile() ([]byte, int64, error) {

	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(math.Pow(2, 14))))
	if err != nil {
		return nil, 0, err
	}
	fileSize := nBig.Int64()
	log.Debug.Printf("Generating file of size %d bytes", fileSize)

	randomBytes := make([]byte, fileSize)
	_, err = rand.Read(randomBytes)
	if err != nil {
		return nil, 0, err
	}

	return randomBytes, fileSize, nil
}
