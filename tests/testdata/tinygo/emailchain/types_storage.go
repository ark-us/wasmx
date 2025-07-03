package main

type FolderState struct {
	Owner   string `json:"owner"`
	Folder  string `json:"folder"`
	LastUid int    `json:"last_uid"`
}

type UidResponse struct {
	LastUid int `json:"last_uid"`
}

type SeqNumResponse struct {
	NextSeqNum int `json:"next_seq_num"`
}
