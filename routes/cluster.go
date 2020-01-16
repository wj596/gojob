/*
 * Copyright 2020-2021 the original author(https://github.com/wj596)
 *
 * <p>
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * </p>
 */
package routes

import (
	"gojob/models"
	"net/http"

	"gojob/internal"

	"github.com/gin-gonic/gin"
)

func getClusterLeaderId(c *gin.Context) {
	c.String(http.StatusOK, internal.GetLeaderId())
}

func joinCluster(c *gin.Context) {
	peerNodeName := c.Param("peer_node_name")
	peerHttpAddr := c.Param("peer_http_addr")
	peerTcpAddr := c.Param("peer_tcp_addr")

	existed, err := models.GetNode(peerNodeName)
	if err != nil {
		node := &models.Node{
			Name:     peerNodeName,
			HttpAddr: peerHttpAddr,
			TcpAddr:  peerTcpAddr,
		}
		internal.InsertNode(node)
	} else {
		if existed.TcpAddr != peerTcpAddr {
			internal.RemovePeer(peerNodeName)
		}
		if existed.HttpAddr != peerHttpAddr || existed.TcpAddr != peerTcpAddr {
			existed.HttpAddr = peerHttpAddr
			existed.TcpAddr = peerTcpAddr
			internal.UpdateNode(existed)
		}
	}

	err = internal.AddVoter(peerNodeName, peerTcpAddr)
	if nil != err {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.String(http.StatusOK, "succeed")
}

func removePeer(c *gin.Context) {
	peerNodeName := c.Param("peer_node_name")
	err := internal.RemovePeer(peerNodeName)
	if nil != err {
		respond500(c, err.Error())
		return
	}
	respondOK(c)
}

func getClusterNodes(c *gin.Context) {
	respondData(c, internal.GetRaftNodeDetails())
}
