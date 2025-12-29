package file_test

import (
	"crypto/rand"
	"fmt"
	"io"
	"math"
	"math/big"
	"testing"
	"time"

	tower_model "github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/errors"
	"github.com/ethanrous/weblens/modules/fs"
	"github.com/ethanrous/weblens/modules/structs"
	"github.com/ethanrous/weblens/services/proxy"
	"github.com/rs/zerolog"
)

func TestFiles(_ *testing.T) {
	// t.Parallel()
	//
	// logger := log.NewZeroLogger()
	// nop := zerolog.Nop()
	//
	// coreCtx, err := test.NewWeblensTestInstance(t.Name(), config.ConfigProvider{
	// 	InitRole: string(tower_model.CoreServerRole),
	// }, &nop)
	//
	// require.NoError(t, err)
	//
	// keys, err := coreServices.AccessService.GetKeysByUser(coreServices.UserService.Get("test-username"))
	// if err != nil {
	// 	logger.Error().Stack().Err(err).Msg("")
	// 	t.FailNow()
	// }
	//
	// coreInstance := coreServices.InstanceService.GetLocal()
	// coreAddress := config.GetProxyAddress(coreServices.Cnf)
	// coreApiKey := keys[0].Key
	//
	// localCoreInstance := &tower_model.Instance{
	// 	TowerID:     coreInstance.ID,
	// 	Name:        coreInstance.Name,
	// 	OutgoingKey: coreApiKey,
	// 	Role:        tower_model.CoreServerRole,
	// 	IsThisTower: false,
	// 	Address:     coreAddress,
	// }
	// owner := coreServices.UserService.Get("test-username")
	// if owner == nil {
	// 	t.Fatalf("No owner")
	// }
	//
	// err = simpleCreate(localCoreInstance, owner)
	// if err != nil {
	// 	logger.Error().Stack().Err(err).Msg("")
	// 	t.FailNow()
	// }
	//
	// err = uploadFile(localCoreInstance, owner, logger)
	// if err != nil {
	// 	logger.Error().Stack().Err(err).Msg("")
	// 	t.FailNow()
	// }
	//
	// err = moveFiles(localCoreInstance, owner)
	// if err != nil {
	// 	logger.Error().Stack().Err(err).Msg("")
	// 	t.FailNow()
	// }
	//
	// ownerHomeInfoRequest := proxy.NewCoreRequest(localCoreInstance, "GET", "/folder/"+owner.HomeID)
	// ownerHomeInfo, err := proxy.CallHomeStruct[structs.FolderInfoResponse](ownerHomeInfoRequest)
	// if err != nil {
	// 	logger.Error().Stack().Err(err).Msg("")
	// 	t.FailNow()
	// }
	//
	// if len(ownerHomeInfo.Children) != 1 || ownerHomeInfo.Children[0].ID != owner.TrashID {
	// 	t.Fatalf("Owner home should be empty: %v", ownerHomeInfo.Children)
	// }
}

func simpleCreate(core *tower_model.Instance, owner *user.User) error {
	createFolderRequest := proxy.NewCoreRequest(core, "POST", "/folder").WithBody(structs.CreateFolderBody{
		ParentFolderID: owner.HomeID,
		NewFolderName:  "test-folder",
	})

	folderInfo, err := proxy.CallHomeStruct[structs.FileInfo](createFolderRequest)
	if err != nil {
		return err
	}

	if folderInfo.ParentID != owner.HomeID {
		return errors.Errorf("Parent folder mismatch %s != %s", folderInfo.ParentID, owner.HomeID)
	}

	getFileRequest := proxy.NewCoreRequest(core, "GET", "/files/"+folderInfo.ID)

	getFolderInfo, err := proxy.CallHomeStruct[structs.FileInfo](getFileRequest)
	if err != nil {
		return err
	}

	folderPath, err := fs.ParsePortable(getFolderInfo.PortablePath)
	if err != nil {
		return err
	}

	if folderPath.Filename() != "test-folder" {
		return errors.Errorf("Folder name mismatch %s != %s", folderPath.Filename(), "test-folder")
	}

	_, err = proxy.NewCoreRequest(core, "DELETE", "/files").WithQuery("ignore_trash", "true").WithBody(structs.FilesListParams{FileIDs: []string{getFolderInfo.ID}}).Call()
	if err != nil {
		return err
	}

	return nil
}

func uploadFile(core *tower_model.Instance, owner *user.User, logger zerolog.Logger) error {
	randomBytes, fileSize, err := makeRandomFile()
	if err != nil {
		return err
	}

	createUploadRequest := proxy.NewCoreRequest(core, "POST", "/upload").WithBody(structs.NewUploadParams{
		RootFolderID: owner.HomeID,
		ChunkSize:    fileSize,
	})

	uploadInfo, err := proxy.CallHomeStruct[structs.NewUploadInfo](createUploadRequest)
	if err != nil {
		return err
	}

	newFileRequest := proxy.NewCoreRequest(core, "POST", "/upload/"+uploadInfo.UploadID).WithBody(structs.NewFilesParams{
		NewFiles: []structs.NewFileParams{
			{
				NewFileName:    "test-file.txt",
				ParentFolderID: owner.HomeID,
				FileSize:       fileSize,
			},
		},
	})

	newFilesInfo, err := proxy.CallHomeStruct[structs.NewFilesInfo](newFileRequest)
	if err != nil {
		return err
	}

	_, err = proxy.NewCoreRequest(core, "PUT", "/upload/"+uploadInfo.UploadID+"/file/"+newFilesInfo.FileIDs[0]).WithHeader("Content-Range", fmt.Sprintf("0-%d/%d", fileSize, fileSize)).WithBodyBytes(randomBytes).Call()
	if err != nil {
		return err
	}

	getFileRequest := proxy.NewCoreRequest(core, "GET", "/files/"+newFilesInfo.FileIDs[0])

	getFileInfo, err := proxy.CallHomeStruct[structs.FileInfo](getFileRequest)
	if err != nil {
		return err
	}

	filePath, err := fs.ParsePortable(getFileInfo.PortablePath)
	if err != nil {
		return err
	}

	if filePath.Filename() != "test-file.txt" {
		return errors.Errorf("File name mismatch %s != %s", filePath.Filename(), "test-file.txt")
	}

	if getFileInfo.ParentID != owner.HomeID {
		return errors.Errorf("Parent folder mismatch %s != %s", getFileInfo.ParentID, owner.HomeID)
	}

	returnedSize := getFileInfo.Size
	timeout := 10

	for returnedSize != fileSize {
		logger.Debug().Func(func(e *zerolog.Event) { e.Msgf("Waiting for file upload to complete") })
		time.Sleep(100 * time.Millisecond)

		timeout--
		if timeout == 0 {
			return errors.Errorf("Did not receive expected file size %d != %d", returnedSize, fileSize)
		}

		getFileRequest := proxy.NewCoreRequest(core, "GET", "/files/"+newFilesInfo.FileIDs[0])

		getFileInfo, err := proxy.CallHomeStruct[structs.FileInfo](getFileRequest)
		if err != nil {
			return err
		}

		returnedSize = getFileInfo.Size
	}

	downloadURL := fmt.Sprintf("/files/%s/download", newFilesInfo.FileIDs[0])

	downloadResponse, err := proxy.NewCoreRequest(core, "GET", downloadURL).Call()
	if err != nil {
		return err
	}

	bodyBytes, err := io.ReadAll(downloadResponse.Body)
	if err != nil {
		return err
	}

	if len(bodyBytes) != int(fileSize) {
		return errors.Errorf("Downloaded file size mismatch %d != %d", len(bodyBytes), fileSize)
	}

	if string(bodyBytes) != string(randomBytes) {
		return errors.Errorf("Downloaded file content mismatch")
	}

	_, err = proxy.NewCoreRequest(core, "DELETE", "/files").WithQuery("ignore_trash", "true").WithBody(structs.FilesListParams{FileIDs: []string{newFilesInfo.FileIDs[0]}}).Call()
	if err != nil {
		return err
	}

	return nil
}

func moveFiles(core *tower_model.Instance, owner *user.User) error {
	folder1, err := makeFolder("top-folder-1", owner.HomeID, core)
	if err != nil {
		return err
	}

	folder2, err := makeFolder("top-folder-2", owner.HomeID, core)
	if err != nil {
		return err
	}

	subFolder, err := makeFolder("sub-folder", folder1.ID, core)
	if err != nil {
		return err
	}

	bytes, fileSize, err := makeRandomFile()
	if err != nil {
		return err
	}

	createUploadRequest := proxy.NewCoreRequest(core, "POST", "/upload").WithBody(structs.NewUploadParams{
		RootFolderID: subFolder.ID,
		ChunkSize:    fileSize,
	})

	uploadInfo, err := proxy.CallHomeStruct[structs.NewUploadInfo](createUploadRequest)
	if err != nil {
		return err
	}

	newFileRequest := proxy.NewCoreRequest(core, "POST", "/upload/"+uploadInfo.UploadID).WithBody(structs.NewFilesParams{
		NewFiles: []structs.NewFileParams{
			{
				NewFileName:    "test-file.txt",
				ParentFolderID: subFolder.ID,
				FileSize:       fileSize,
			},
		},
	})

	newFilesInfo, err := proxy.CallHomeStruct[structs.NewFilesInfo](newFileRequest)
	if err != nil {
		return err
	}

	_, err = proxy.NewCoreRequest(core, "PUT", "/upload/"+uploadInfo.UploadID+"/file/"+newFilesInfo.FileIDs[0]).WithHeader("Content-Range", fmt.Sprintf("0-%d/%d", fileSize, fileSize)).WithBodyBytes(bytes).Call()
	if err != nil {
		return err
	}

	// Wait for upload to complete
	_, err = proxy.NewCoreRequest(core, "GET", "/upload/"+uploadInfo.UploadID).Call()
	if err != nil {
		return err
	}

	_, err = proxy.NewCoreRequest(core, "PATCH", "/files/").WithBody(structs.MoveFilesParams{
		NewParentID: folder2.ID,
		Files:       []string{subFolder.ID},
	}).Call()
	if err != nil {
		return err
	}

	getFolder1Request := proxy.NewCoreRequest(core, "GET", "/folder/"+folder1.ID)

	folder1Info, err := proxy.CallHomeStruct[structs.FolderInfoResponse](getFolder1Request)
	if err != nil {
		return err
	}

	if len(folder1Info.Children) != 0 {
		return errors.Errorf("Folder 1 should be empty")
	}

	getFolder2Request := proxy.NewCoreRequest(core, "GET", "/folder/"+folder2.ID)

	folder2Info, err := proxy.CallHomeStruct[structs.FolderInfoResponse](getFolder2Request)
	if err != nil {
		return err
	}

	if len(folder2Info.Children) != 1 {
		return errors.Errorf("Folder 2 should have 1 child")
	}

	if folder2Info.Children[0].ID != subFolder.ID {
		return errors.Errorf("Folder 2 should have sub-folder as child")
	}

	_, err = proxy.NewCoreRequest(core, "DELETE", "/files").WithQuery("ignore_trash", "true").WithBody(structs.FilesListParams{FileIDs: []string{folder1.ID, folder2.ID}}).Call()
	if err != nil {
		return err
	}

	return nil
}

func makeFolder(name string, parentFolderID string, core *tower_model.Instance) (structs.FileInfo, error) {
	createFolderRequest := proxy.NewCoreRequest(core, "POST", "/folder").WithBody(structs.CreateFolderBody{
		ParentFolderID: parentFolderID,
		NewFolderName:  name,
	})

	folderInfo, err := proxy.CallHomeStruct[structs.FileInfo](createFolderRequest)
	if err != nil {
		return structs.FileInfo{}, err
	}

	return folderInfo, nil
}

func makeRandomFile() ([]byte, int64, error) {
	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(math.Pow(2, 14))))
	if err != nil {
		return nil, 0, err
	}

	fileSize := nBig.Int64()

	randomBytes := make([]byte, fileSize)

	_, err = rand.Read(randomBytes)
	if err != nil {
		return nil, 0, err
	}

	return randomBytes, fileSize, nil
}
