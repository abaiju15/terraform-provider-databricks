package workspace

import (
	"context"
	"net/http"
	"testing"

	"github.com/databricks/databricks-sdk-go/apierr"
	"github.com/databricks/terraform-provider-databricks/qa"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceNotebookRead(t *testing.T) {
	path := "/test/path.py"
	objectID := 12345
	d, err := qa.ResourceFixture{
		Fixtures: []qa.HTTPFixture{
			{
				Method:   http.MethodGet,
				Resource: "/api/2.0/workspace/get-status?path=%2Ftest%2Fpath.py",
				Response: ObjectStatus{
					ObjectID:   int64(objectID),
					ObjectType: Notebook,
					Path:       path,
					Language:   "PYTHON",
				},
			},
		},
		Resource: ResourceNotebook(),
		Read:     true,
		New:      true,
		ID:       path,
	}.Apply(t)
	assert.NoError(t, err)
	assert.Equal(t, path, d.Id())
	assert.Equal(t, path, d.Get("path"))
	assert.Equal(t, "PYTHON", d.Get("language"))
	assert.Equal(t, objectID, d.Get("object_id"))
}

func TestResourceNotebookDelete(t *testing.T) {
	path := "/test/path.py"
	d, err := qa.ResourceFixture{
		Fixtures: []qa.HTTPFixture{
			{
				Method:          http.MethodPost,
				Resource:        "/api/2.0/workspace/delete",
				Status:          http.StatusOK,
				ExpectedRequest: DeletePath{Path: path, Recursive: true},
			},
		},
		Resource: ResourceNotebook(),
		Delete:   true,
		ID:       path,
	}.Apply(t)
	assert.NoError(t, err)
	assert.Equal(t, path, d.Id())
}

func TestResourceNotebookRead_NotFound(t *testing.T) {
	qa.ResourceFixture{
		Fixtures: []qa.HTTPFixture{
			{ // read log output for correct url...
				Method:   "GET",
				Resource: "/api/2.0/workspace/get-status?path=%2Ftest%2Fpath",
				Response: apierr.APIErrorBody{
					ErrorCode: "NOT_FOUND",
					Message:   "Item not found",
				},
				Status: 404,
			},
		},
		Resource: ResourceNotebook(),
		Read:     true,
		Removed:  true,
		ID:       "/test/path",
	}.ApplyNoError(t)
}

func TestResourceNotebookRead_Error(t *testing.T) {
	d, err := qa.ResourceFixture{
		Fixtures: []qa.HTTPFixture{
			{
				Method:   "GET",
				Resource: "/api/2.0/workspace/get-status?path=%2Ftest%2Fpath",
				Response: apierr.APIErrorBody{
					ErrorCode: "INVALID_REQUEST",
					Message:   "Internal error happened",
				},
				Status: 400,
			},
		},
		Resource: ResourceNotebook(),
		Read:     true,
		ID:       "/test/path",
	}.Apply(t)
	qa.AssertErrorStartsWith(t, err, "Internal error happened")
	assert.Equal(t, "/test/path", d.Id(), "Id should not be empty for error reads")
}

func TestResourceNotebookCreate(t *testing.T) {
	d, err := qa.ResourceFixture{
		Fixtures: []qa.HTTPFixture{
			{
				Method:   "POST",
				Resource: "/api/2.0/workspace/mkdirs",
				ExpectedRequest: map[string]string{
					"path": "/foo",
				},
			},
			{
				Method:   http.MethodPost,
				Resource: "/api/2.0/workspace/import",
				ExpectedRequest: ImportPath{
					Content:   "YWJjCg==",
					Path:      "/foo/path.py",
					Language:  "PYTHON",
					Overwrite: true,
					Format:    "SOURCE",
				},
			},
			{
				Method:   http.MethodGet,
				Resource: "/api/2.0/workspace/export?format=SOURCE&path=%2Ffoo%2Fpath.py",
				Response: ExportPath{
					Content: "YWJjCg==",
				},
			},
			{
				Method:   http.MethodGet,
				Resource: "/api/2.0/workspace/get-status?path=%2Ffoo%2Fpath.py",
				Response: ObjectStatus{
					ObjectID:   4567,
					ObjectType: "NOTEBOOK",
					Path:       "/foo/path.py",
					Language:   "PYTHON",
				},
			},
		},
		Resource: ResourceNotebook(),
		State: map[string]any{
			"content_base64": "YWJjCg==",
			"language":       "PYTHON",
			"path":           "/foo/path.py",
		},
		Create: true,
	}.Apply(t)
	assert.NoError(t, err)
	assert.Equal(t, "/foo/path.py", d.Id())
}

func TestResourceNotebookCreateSource_Jupyter(t *testing.T) {
	d, err := qa.ResourceFixture{
		Fixtures: []qa.HTTPFixture{
			{
				Method:   http.MethodPost,
				Resource: "/api/2.0/workspace/import",
				ExpectedRequest: ImportPath{
					Content: "eyJjZWxscyI6W3siY2VsbF90eXBlIjoiY29kZSIsInNvdXJjZSI6WyJwc" +
						"mludChcImhlbGxvIHdvcmxkXCIpIl0sIm1ldGFkYXRhIjp7fSwib3V0cHV" +
						"0cyI6W10sImV4ZWN1dGlvbl9jb3VudCI6MX0seyJjZWxsX3R5cGUiOiJjb" +
						"2RlIiwic291cmNlIjpbInByaW50KFwiaG93IGFyZSB5b3VcIikiXSwibWV" +
						"0YWRhdGEiOnt9LCJvdXRwdXRzIjpbeyJtZXRhZGF0YSI6e30sIm91dHB1d" +
						"F90eXBlIjoiZGlzcGxheV9kYXRhIiwiZGF0YSI6eyJ0ZXh0L2h0bWwiOls" +
						"iPHN0eWxlIHNjb3BlZD5cbiAgLmFuc2lvdXQge1xuICAgIGRpc3BsYXk6I" +
						"GJsb2NrO1xuICAgIHVuaWNvZGUtYmlkaTogZW1iZWQ7XG4gICAgd2hpdGU" +
						"tc3BhY2U6IHByZS13cmFwO1xuICAgIHdvcmQtd3JhcDogYnJlYWstd29yZ" +
						"DtcbiAgICB3b3JkLWJyZWFrOiBicmVhay1hbGw7XG4gICAgZm9udC1mYW1" +
						"pbHk6IFwiU291cmNlIENvZGUgUHJvXCIsIFwiTWVubG9cIiwgbW9ub3NwY" +
						"WNlOztcbiAgICBmb250LXNpemU6IDEzcHg7XG4gICAgY29sb3I6ICM1NTU" +
						"7XG4gICAgbWFyZ2luLWxlZnQ6IDRweDtcbiAgICBsaW5lLWhlaWdodDogM" +
						"TlweDtcbiAgfVxuPC9zdHlsZT5cbjxkaXYgY2xhc3M9XCJhbnNpb3V0XCI" +
						"+aG93IGFyZSB5b3VcbjwvZGl2PiJdfX1dLCJleGVjdXRpb25fY291bnQiO" +
						"jJ9LHsiY2VsbF90eXBlIjoiY29kZSIsInNvdXJjZSI6WyIiXSwibWV0YWR" +
						"hdGEiOnt9LCJvdXRwdXRzIjpbXSwiZXhlY3V0aW9uX2NvdW50IjozfV0sI" +
						"m1ldGFkYXRhIjp7Im5hbWUiOiJ0ZXN0X2p1cHl0ZXIiLCJub3RlYm9va0l" +
						"kIjoxMjc1OTg0MjQzMjkzMDI4fSwibmJmb3JtYXQiOjQsIm5iZm9ybWF0X" +
						"21pbm9yIjowfQo=",
					Path:      "/Mars",
					Language:  "",
					Overwrite: true,
					Format:    "JUPYTER",
				},
			},
			{
				Method:   http.MethodGet,
				Resource: "/api/2.0/workspace/get-status?path=%2FMars",
				Response: ObjectStatus{
					ObjectID:   4567,
					ObjectType: "NOTEBOOK",
					Path:       "/Mars",
				},
			},
		},
		Resource: ResourceNotebook(),
		State: map[string]any{
			"source": "acceptance/testdata/tf-test-jupyter.ipynb",
			"path":   "/Mars",
		},
		Create: true,
	}.Apply(t)
	assert.NoError(t, err)
	assert.Equal(t, "/Mars", d.Id())
}

func TestResourceNotebookCreateSource(t *testing.T) {
	d, err := qa.ResourceFixture{
		Fixtures: []qa.HTTPFixture{
			{
				Method:   http.MethodPost,
				Resource: "/api/2.0/workspace/import",
				ExpectedRequest: ImportPath{
					Content: "LS0gRGF0YWJyaWNrcyBub3RlYm9vayBzb3VyY2UKU0VMRUNUIDEwKjIwC" +
						"gotLSBDT01NQU5EIC0tLS0tLS0tLS0KClNFTEVDVCAyMCoxMDAKCi0tIE" +
						"NPTU1BTkQgLS0tLS0tLS0tLQoKCg==",
					Path:      "/Dashboard",
					Language:  "SQL",
					Overwrite: true,
					Format:    "SOURCE",
				},
			},
			{
				Method:   http.MethodGet,
				Resource: "/api/2.0/workspace/get-status?path=%2FDashboard",
				Response: ObjectStatus{
					ObjectID:   4567,
					ObjectType: "NOTEBOOK",
					Path:       "/Dashboard",
					Language:   "SQL",
				},
			},
		},
		Resource: ResourceNotebook(),
		State: map[string]any{
			"source": "acceptance/testdata/tf-test-sql.sql",
			"path":   "/Dashboard",
		},
		Create: true,
	}.Apply(t)
	assert.NoError(t, err)
	assert.Equal(t, "/Dashboard", d.Id())
}

func TestResourceNotebookCreate_Error(t *testing.T) {
	d, err := qa.ResourceFixture{
		Fixtures: []qa.HTTPFixture{
			{
				Method:   http.MethodPost,
				Resource: "/api/2.0/workspace/import",
				Response: apierr.APIErrorBody{
					ErrorCode: "INVALID_REQUEST",
					Message:   "Internal error happened",
				},
				Status: 400,
			},
		},
		Resource: ResourceNotebook(),
		State: map[string]any{
			"content_base64": "YWJjCg==",
			"language":       "R",
			"path":           "/path.py",
		},
		Create: true,
	}.Apply(t)
	qa.AssertErrorStartsWith(t, err, "Internal error happened")
	assert.Equal(t, "", d.Id(), "Id should be empty for error creates")
}

func TestResourceNotebookDelete_Error(t *testing.T) {
	d, err := qa.ResourceFixture{
		Fixtures: []qa.HTTPFixture{
			{
				Method:   "POST",
				Resource: "/api/2.0/workspace/delete",
				Response: apierr.APIErrorBody{
					ErrorCode: "INVALID_REQUEST",
					Message:   "Internal error happened",
				},
				Status: 400,
			},
		},
		Resource: ResourceNotebook(),
		Delete:   true,
		ID:       "abc",
	}.Apply(t)
	qa.AssertErrorStartsWith(t, err, "Internal error happened")
	assert.Equal(t, "abc", d.Id())
}

func TestResourceNotebookUpdate(t *testing.T) {
	qa.ResourceFixture{
		Fixtures: []qa.HTTPFixture{
			{
				Method:   "POST",
				Resource: "/api/2.0/workspace/import",
				ExpectedRequest: ImportPath{
					Format:    "SOURCE",
					Overwrite: true,
					Content:   "YWJjCg==",
					Path:      "abc",
					Language:  "R",
				},
			},
			{
				Method:   http.MethodGet,
				Resource: "/api/2.0/workspace/get-status?path=abc",
				Response: ObjectStatus{
					ObjectID:   4567,
					ObjectType: "NOTEBOOK",
					Path:       "abc",
					Language:   "R",
				},
			},
		},
		Resource: ResourceNotebook(),
		State: map[string]any{
			"content_base64": "YWJjCg==",
			"language":       "R",
			"path":           "/path.py",
		},
		ID:          "abc",
		RequiresNew: true,
		Update:      true,
	}.ApplyNoError(t)
}

func TestResourceNotebookUpdate_DBC(t *testing.T) {
	qa.ResourceFixture{
		Fixtures: []qa.HTTPFixture{
			{
				Method:   "POST",
				Resource: "/api/2.0/workspace/delete",
				ExpectedRequest: DeletePath{
					Recursive: true,
					Path:      "abc",
				},
			},
			{
				Method:   "POST",
				Resource: "/api/2.0/workspace/import",
				ExpectedRequest: ImportPath{
					Format:  "DBC",
					Content: "YWJjCg==",
					Path:    "abc",
				},
			},
			{
				Method:   http.MethodGet,
				Resource: "/api/2.0/workspace/get-status?path=abc",
				Response: ObjectStatus{
					ObjectID:   4567,
					ObjectType: Directory,
					Path:       "abc",
				},
			},
		},
		Resource: ResourceNotebook(),
		State: map[string]any{
			"content_base64": "YWJjCg==",

			// technically language is not needed, but makes the test simpler
			"language": "PYTHON",
			"format":   "DBC",
			"path":     "/path.py",
		},
		ID:          "abc",
		RequiresNew: true,
		Update:      true,
	}.ApplyNoError(t)
}

func TestNotebookLanguageSuppressSourceDiff(t *testing.T) {
	r := ResourceNotebook()
	d := r.TestResourceData()
	d.Set("source", "this.PY")
	suppress := r.Schema["language"].DiffSuppressFunc
	assert.True(t, suppress("language", Python, Python, d))
}

func TestListDirectories(t *testing.T) {
	client, server, err := qa.HttpFixtureClient(t, []qa.HTTPFixture{
		{
			Method:   http.MethodGet,
			Resource: "/api/2.0/workspace/list?path=%2F",
			Response: ObjectList{
				Objects: []ObjectStatus{
					{
						ObjectID:   1,
						Path:       "/Parent",
						ObjectType: "DIRECTORY",
					},
				},
			},
		},
		{
			Method:   "GET",
			Resource: "/api/2.0/workspace/list?path=%2FParent",
			Response: ObjectList{
				Objects: []ObjectStatus{
					{
						ObjectID:   11,
						ObjectType: Directory,
						Path:       "/Parent/Kid1",
					},
					{
						ObjectID:   12,
						ObjectType: Directory,
						Path:       "/Parent/Kid2",
					},
				},
			},
		},
		{
			Method:   "GET",
			Resource: "/api/2.0/workspace/list?path=%2FParent%2FKid1",
			Response: ObjectList{
				Objects: []ObjectStatus{
					{
						ObjectID:   111,
						ObjectType: Directory,
						Path:       "/Parent/Kid1/GrandKid1",
					},
					{
						ObjectID:   112,
						ObjectType: Notebook,
						Path:       "/Parent/Kid1/GrandKid2",
						Language:   Python,
					},
				},
			},
		},

		{
			Method:   "GET",
			Resource: "/api/2.0/workspace/list?path=%2FParent%2FKid2",
			Response: ObjectList{
				Objects: []ObjectStatus{
					{},
				},
			},
		},
		{
			Method:   "GET",
			Resource: "/api/2.0/workspace/list?path=%2FParent%2FKid1%2FGrandKid1",
			Response: ObjectList{
				Objects: []ObjectStatus{
					{},
				},
			},
		},
	})
	defer server.Close()
	require.NoError(t, err)

	ctx := context.Background()
	notebooksAPI := NewNotebooksAPI(ctx, client)

	directoryList, err := notebooksAPI.ListDirectories("/", true)
	var paths []string
	for _, directory := range directoryList {
		paths = append(paths, directory.Path)
	}
	assert.NoError(t, err, err)
	assert.NotNil(t, directoryList)
	assert.Len(t, directoryList, 4)
	assert.Contains(t, paths, "/Parent")
	assert.Contains(t, paths, "/Parent/Kid1")
	assert.Contains(t, paths, "/Parent/Kid2")
	assert.Contains(t, paths, "/Parent/Kid1/GrandKid1")
}

func TestListDirectories_NoneRecursive(t *testing.T) {
	client, server, err := qa.HttpFixtureClient(t, []qa.HTTPFixture{
		{
			Method:   http.MethodGet,
			Resource: "/api/2.0/workspace/list?path=%2F",
			Response: ObjectList{
				Objects: []ObjectStatus{
					{
						ObjectID:   1,
						Path:       "/Parent",
						ObjectType: "DIRECTORY",
					},
				},
			},
		},
		{
			Method:   "GET",
			Resource: "/api/2.0/workspace/list?path=%2FParent",
			Response: ObjectList{
				Objects: []ObjectStatus{
					{
						ObjectID:   11,
						ObjectType: Directory,
						Path:       "/Parent/Kid1",
					},
					{
						ObjectID:   12,
						ObjectType: Directory,
						Path:       "/Parent/Kid2",
					},
				},
			},
		},
		{
			Method:   "GET",
			Resource: "/api/2.0/workspace/list?path=%2FParent%2FKid1",
			Response: ObjectList{
				Objects: []ObjectStatus{
					{
						ObjectID:   111,
						ObjectType: Directory,
						Path:       "/Parent/Kid1/GrandKid1",
					},
					{
						ObjectID:   112,
						ObjectType: Notebook,
						Path:       "/Parent/Kid1/GrandKid2",
						Language:   Python,
					},
				},
			},
		},
		{
			Method:   "GET",
			Resource: "/api/2.0/workspace/list?path=%2FParent%2FKid2",
			Response: ObjectList{
				Objects: []ObjectStatus{
					{},
				},
			},
		},
		{
			Method:   "GET",
			Resource: "/api/2.0/workspace/list?path=%2FParent%2FKid1%2FGrandKid1",
			Response: ObjectList{
				Objects: []ObjectStatus{
					{},
				},
			},
		},
	})
	defer server.Close()
	require.NoError(t, err)
	ctx := context.Background()
	notebooksAPI := NewNotebooksAPI(ctx, client)
	directoryList_not_recursive, err := notebooksAPI.ListDirectories("/", false)
	var paths []string
	for _, directory := range directoryList_not_recursive {
		paths = append(paths, directory.Path)
	}
	assert.NoError(t, err, err)
	assert.NotNil(t, directoryList_not_recursive)
	assert.Len(t, directoryList_not_recursive, 1)
	assert.Contains(t, paths, "/Parent")
}

func TestListDirectoriesRecursive_Error(t *testing.T) {
	client, server, err := qa.HttpFixtureClient(t, []qa.HTTPFixture{
		{
			Method:   http.MethodGet,
			Resource: "/api/2.0/workspace/list?path=%2F",
			Response: ObjectList{
				Objects: []ObjectStatus{
					{
						ObjectID:   1,
						Path:       "/Parent",
						ObjectType: "DIRECTORY",
					},
				},
			},
		},
		{
			Method:   http.MethodGet,
			Resource: "/api/2.0/workspace/list?path=%2FParent",
			Response: apierr.APIErrorBody{
				ErrorCode: "Internal Error",
				Message:   "Internal Error",
			},
			Status: 400,
		},
	})
	defer server.Close()
	require.NoError(t, err)
	ctx := context.Background()
	notebooksAPI := NewNotebooksAPI(ctx, client)
	directoryList, err := notebooksAPI.ListDirectories("/", true)
	assert.Error(t, err, err)
	assert.Nil(t, directoryList)
}
