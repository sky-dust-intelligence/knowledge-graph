package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/oseducation/knowledge-graph/model"
)

const (
	defaultNodePage    = -1
	defaultNodePerPage = -1
)

func (apiObj *API) initNode() {
	apiObj.Nodes = apiObj.APIRoot.Group("/nodes")

	apiObj.Nodes.GET("/", apiObj.jwtMiddleware.MiddlewareFunc(), getNodes)
	apiObj.Nodes.POST("/", apiObj.jwtMiddleware.MiddlewareFunc(), createNode)
	apiObj.Nodes.PUT("/", apiObj.jwtMiddleware.MiddlewareFunc(), updateNode)
	apiObj.Nodes.DELETE("/", apiObj.jwtMiddleware.MiddlewareFunc(), deleteNode)
}

func createNode(c *gin.Context) {
	node, err := model.NodeFromJSON(c.Request.Body)
	if err != nil {
		responseFormat(c, http.StatusBadRequest, "Invalid or missing `node` in the request body", nil)
		return
	}

	a, err := getApp(c)
	if err != nil {
		responseFormat(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	authorID := c.GetString(identityKey)
	if !a.HasPermissionToManageNodes(authorID) {
		responseFormat(c, http.StatusForbidden, "No permission for this action", nil)
		return
	}

	rnode, err := a.CreateNode(node)
	if err != nil {
		responseFormat(c, http.StatusInternalServerError, "Error while creating node", nil)
		return
	}
	responseFormat(c, http.StatusCreated, "Node created", rnode)
}

func getNodes(c *gin.Context) {
	termInName := c.DefaultQuery("term_in_name", "")
	termInDescription := c.DefaultQuery("term_in_description", "")
	page, err := strconv.Atoi(c.DefaultQuery("page", strconv.Itoa(defaultNodePage)))
	if err != nil {
		page = defaultUserPage
	}
	perPage, err := strconv.Atoi(c.DefaultQuery("per_page", strconv.Itoa(defaultNodePerPage)))
	if err != nil {
		perPage = defaultUserPerPage
	}

	a, err := getApp(c)
	if err != nil {
		responseFormat(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	authorID := c.GetString(identityKey)
	if !a.HasPermissionToManageNodes(authorID) {
		responseFormat(c, http.StatusForbidden, "No permission for this action", nil)
		return
	}

	options := &model.NodeGetOptions{}
	model.ComposeNodeOptions(
		model.TermInName(termInName),
		model.TermInDescription(termInDescription),
		model.NodePage(page),
		model.NodePerPage(perPage))(options)
	nodes, err := a.GetNodes(options)
	if err != nil {
		responseFormat(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	responseFormat(c, http.StatusOK, "", nodes)
}

func updateNode(c *gin.Context) {
	updatedNode, err := model.NodeFromJSON(c.Request.Body)
	if err != nil {
		responseFormat(c, http.StatusBadRequest, "Invalid or missing `node` in the request body", nil)
		return
	}
	a, err := getApp(c)
	if err != nil {
		responseFormat(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	authorID := c.GetString(identityKey)
	if !a.HasPermissionToManageNodes(authorID) {
		responseFormat(c, http.StatusForbidden, "No permission for this action", nil)
		return
	}

	err = a.UpdateNode(updatedNode)
	if err != nil {
		responseFormat(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	responseFormat(c, http.StatusOK, "Node updated", nil)
}

func deleteNode(c *gin.Context) {
	nodeID := c.Query("node_id")
	if nodeID == "" {
		responseFormat(c, http.StatusBadRequest, "missing node_id", nil)
		return
	}
	a, err := getApp(c)
	if err != nil {
		responseFormat(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	authorID := c.GetString(identityKey)
	if !a.HasPermissionToManageNodes(authorID) {
		responseFormat(c, http.StatusForbidden, "No permission for this action", nil)
		return
	}

	node, err := a.Store.Node().Get(nodeID)
	if err != nil {
		responseFormat(c, http.StatusBadRequest, "unknown node", nil)
		return
	}

	if err = a.DeleteNode(node); err != nil {
		responseFormat(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	responseFormat(c, http.StatusOK, "Node deleted", nil)
}