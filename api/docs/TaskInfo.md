# TaskInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Completed** | **bool** |  | 
**State** | **string** |  | 
**CompletedChildTasks** | Pointer to **int32** |  | [optional] 
**Error** | Pointer to **string** |  | [optional] 
**FinishTime** | Pointer to **string** |  | [optional] 
**JobName** | **string** |  | 
**Metadata** | Pointer to **map[string]interface{}** |  | [optional] 
**ParentTaskID** | Pointer to **string** |  | [optional] 
**Progress** | **int32** |  | 
**QueueTime** | Pointer to **string** |  | [optional] 
**Result** | Pointer to **map[string]interface{}** |  | [optional] 
**StartTime** | Pointer to **string** |  | [optional] 
**Status** | **string** |  | 
**TaskID** | **string** |  | 
**TotalChildTasks** | Pointer to **int32** |  | [optional] 
**WorkerID** | **int32** |  | 

## Methods

### NewTaskInfo

`func NewTaskInfo(completed bool, state string, jobName string, progress int32, status string, taskID string, workerID int32, ) *TaskInfo`

NewTaskInfo instantiates a new TaskInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTaskInfoWithDefaults

`func NewTaskInfoWithDefaults() *TaskInfo`

NewTaskInfoWithDefaults instantiates a new TaskInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCompleted

`func (o *TaskInfo) GetCompleted() bool`

GetCompleted returns the Completed field if non-nil, zero value otherwise.

### GetCompletedOk

`func (o *TaskInfo) GetCompletedOk() (*bool, bool)`

GetCompletedOk returns a tuple with the Completed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompleted

`func (o *TaskInfo) SetCompleted(v bool)`

SetCompleted sets Completed field to given value.


### GetState

`func (o *TaskInfo) GetState() string`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *TaskInfo) GetStateOk() (*string, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *TaskInfo) SetState(v string)`

SetState sets State field to given value.


### GetCompletedChildTasks

`func (o *TaskInfo) GetCompletedChildTasks() int32`

GetCompletedChildTasks returns the CompletedChildTasks field if non-nil, zero value otherwise.

### GetCompletedChildTasksOk

`func (o *TaskInfo) GetCompletedChildTasksOk() (*int32, bool)`

GetCompletedChildTasksOk returns a tuple with the CompletedChildTasks field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompletedChildTasks

`func (o *TaskInfo) SetCompletedChildTasks(v int32)`

SetCompletedChildTasks sets CompletedChildTasks field to given value.

### HasCompletedChildTasks

`func (o *TaskInfo) HasCompletedChildTasks() bool`

HasCompletedChildTasks returns a boolean if a field has been set.

### GetError

`func (o *TaskInfo) GetError() string`

GetError returns the Error field if non-nil, zero value otherwise.

### GetErrorOk

`func (o *TaskInfo) GetErrorOk() (*string, bool)`

GetErrorOk returns a tuple with the Error field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetError

`func (o *TaskInfo) SetError(v string)`

SetError sets Error field to given value.

### HasError

`func (o *TaskInfo) HasError() bool`

HasError returns a boolean if a field has been set.

### GetFinishTime

`func (o *TaskInfo) GetFinishTime() string`

GetFinishTime returns the FinishTime field if non-nil, zero value otherwise.

### GetFinishTimeOk

`func (o *TaskInfo) GetFinishTimeOk() (*string, bool)`

GetFinishTimeOk returns a tuple with the FinishTime field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFinishTime

`func (o *TaskInfo) SetFinishTime(v string)`

SetFinishTime sets FinishTime field to given value.

### HasFinishTime

`func (o *TaskInfo) HasFinishTime() bool`

HasFinishTime returns a boolean if a field has been set.

### GetJobName

`func (o *TaskInfo) GetJobName() string`

GetJobName returns the JobName field if non-nil, zero value otherwise.

### GetJobNameOk

`func (o *TaskInfo) GetJobNameOk() (*string, bool)`

GetJobNameOk returns a tuple with the JobName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetJobName

`func (o *TaskInfo) SetJobName(v string)`

SetJobName sets JobName field to given value.


### GetMetadata

`func (o *TaskInfo) GetMetadata() map[string]interface{}`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *TaskInfo) GetMetadataOk() (*map[string]interface{}, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *TaskInfo) SetMetadata(v map[string]interface{})`

SetMetadata sets Metadata field to given value.

### HasMetadata

`func (o *TaskInfo) HasMetadata() bool`

HasMetadata returns a boolean if a field has been set.

### GetParentTaskID

`func (o *TaskInfo) GetParentTaskID() string`

GetParentTaskID returns the ParentTaskID field if non-nil, zero value otherwise.

### GetParentTaskIDOk

`func (o *TaskInfo) GetParentTaskIDOk() (*string, bool)`

GetParentTaskIDOk returns a tuple with the ParentTaskID field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParentTaskID

`func (o *TaskInfo) SetParentTaskID(v string)`

SetParentTaskID sets ParentTaskID field to given value.

### HasParentTaskID

`func (o *TaskInfo) HasParentTaskID() bool`

HasParentTaskID returns a boolean if a field has been set.

### GetProgress

`func (o *TaskInfo) GetProgress() int32`

GetProgress returns the Progress field if non-nil, zero value otherwise.

### GetProgressOk

`func (o *TaskInfo) GetProgressOk() (*int32, bool)`

GetProgressOk returns a tuple with the Progress field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProgress

`func (o *TaskInfo) SetProgress(v int32)`

SetProgress sets Progress field to given value.


### GetQueueTime

`func (o *TaskInfo) GetQueueTime() string`

GetQueueTime returns the QueueTime field if non-nil, zero value otherwise.

### GetQueueTimeOk

`func (o *TaskInfo) GetQueueTimeOk() (*string, bool)`

GetQueueTimeOk returns a tuple with the QueueTime field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetQueueTime

`func (o *TaskInfo) SetQueueTime(v string)`

SetQueueTime sets QueueTime field to given value.

### HasQueueTime

`func (o *TaskInfo) HasQueueTime() bool`

HasQueueTime returns a boolean if a field has been set.

### GetResult

`func (o *TaskInfo) GetResult() map[string]interface{}`

GetResult returns the Result field if non-nil, zero value otherwise.

### GetResultOk

`func (o *TaskInfo) GetResultOk() (*map[string]interface{}, bool)`

GetResultOk returns a tuple with the Result field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResult

`func (o *TaskInfo) SetResult(v map[string]interface{})`

SetResult sets Result field to given value.

### HasResult

`func (o *TaskInfo) HasResult() bool`

HasResult returns a boolean if a field has been set.

### GetStartTime

`func (o *TaskInfo) GetStartTime() string`

GetStartTime returns the StartTime field if non-nil, zero value otherwise.

### GetStartTimeOk

`func (o *TaskInfo) GetStartTimeOk() (*string, bool)`

GetStartTimeOk returns a tuple with the StartTime field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStartTime

`func (o *TaskInfo) SetStartTime(v string)`

SetStartTime sets StartTime field to given value.

### HasStartTime

`func (o *TaskInfo) HasStartTime() bool`

HasStartTime returns a boolean if a field has been set.

### GetStatus

`func (o *TaskInfo) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *TaskInfo) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *TaskInfo) SetStatus(v string)`

SetStatus sets Status field to given value.


### GetTaskID

`func (o *TaskInfo) GetTaskID() string`

GetTaskID returns the TaskID field if non-nil, zero value otherwise.

### GetTaskIDOk

`func (o *TaskInfo) GetTaskIDOk() (*string, bool)`

GetTaskIDOk returns a tuple with the TaskID field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTaskID

`func (o *TaskInfo) SetTaskID(v string)`

SetTaskID sets TaskID field to given value.


### GetTotalChildTasks

`func (o *TaskInfo) GetTotalChildTasks() int32`

GetTotalChildTasks returns the TotalChildTasks field if non-nil, zero value otherwise.

### GetTotalChildTasksOk

`func (o *TaskInfo) GetTotalChildTasksOk() (*int32, bool)`

GetTotalChildTasksOk returns a tuple with the TotalChildTasks field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalChildTasks

`func (o *TaskInfo) SetTotalChildTasks(v int32)`

SetTotalChildTasks sets TotalChildTasks field to given value.

### HasTotalChildTasks

`func (o *TaskInfo) HasTotalChildTasks() bool`

HasTotalChildTasks returns a boolean if a field has been set.

### GetWorkerID

`func (o *TaskInfo) GetWorkerID() int32`

GetWorkerID returns the WorkerID field if non-nil, zero value otherwise.

### GetWorkerIDOk

`func (o *TaskInfo) GetWorkerIDOk() (*int32, bool)`

GetWorkerIDOk returns a tuple with the WorkerID field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkerID

`func (o *TaskInfo) SetWorkerID(v int32)`

SetWorkerID sets WorkerID field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


