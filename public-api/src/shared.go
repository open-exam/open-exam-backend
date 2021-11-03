package main

type Scope struct {
	Scope uint64 `json:"scope" binding:"required"`
	ScopeType uint32 `json:"scopeType" binding:"required"`
}