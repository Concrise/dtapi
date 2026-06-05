package shared

import (
	"errors"

	"ds2api/internal/config"
	"ds2api/internal/promptcompat"
)

const ExpertFilesUnsupportedMessage = "DeepSeek Pro/expert mode does not support file uploads or file references."

var ErrExpertFilesUnsupported = errors.New(ExpertFilesUnsupportedMessage)

func IsExpertModel(model string) bool {
	modelType, ok := config.GetModelType(model)
	return ok && modelType == "expert"
}

func ValidateExpertFileRefs(stdReq promptcompat.StandardRequest) error {
	if !IsExpertModel(stdReq.ResolvedModel) || len(stdReq.RefFileIDs) == 0 {
		return nil
	}
	return ErrExpertFilesUnsupported
}
