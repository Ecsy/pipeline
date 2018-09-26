// Copyright © 2018 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package backups

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/banzaicloud/pipeline/api/ark/common"
	"github.com/banzaicloud/pipeline/internal/ark/api"
	"github.com/banzaicloud/pipeline/internal/platform/gin/correlationid"
)

// Delete deletes an ARK backup
func Delete(c *gin.Context) {
	backupName := c.Param("name")

	logger := correlationid.Logger(common.Log, c).WithField("backup", backupName)
	logger.Info("deleting backup")

	err := common.GetARKService(c.Request).GetClusterBackupsService().DeleteByName(backupName)
	if err != nil {
		err = errors.Wrap(err, "error deleting backup")
		logger.Error(err.Error())
		common.ErrorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, &api.DeleteBackupResponse{
		Name:   backupName,
		Status: http.StatusOK,
	})
}
