# TaskInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Completed** | **bool** |  | 
**JobName** | **string** |  | 
**Progress** | **int32** |  | 
**Result** | Pointer to **map[string]interface{}** |  | [optional] 
**StartTime** | Pointer to **string** |  | [optional] 
**Status** | **string** |  | 
**TaskID** | **string** |  | 
**WorkerID** | **int32** |  | 

## Methods

### NewTaskInfo

`func NewTaskInfo(completed bool, jobName string, progress int32, status string, taskID string, workerID int32, ) *TaskInfo`

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


