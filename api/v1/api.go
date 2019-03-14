package v1

import (
	"fmt"
	"log"
	"os"
	"strings"

	loads "github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"
	flags "github.com/jessevdk/go-flags"
	"github.com/mxinden/promql-server/api/v1/models"
	"github.com/mxinden/promql-server/api/v1/restapi"
	"github.com/mxinden/promql-server/api/v1/restapi/operations"
	"github.com/prometheus/prometheus/promql"
)

func Serve() {
	swaggerSpec, err := loads.Embedded(restapi.SwaggerJSON, restapi.FlatSwaggerJSON)
	if err != nil {
		log.Fatalln(err)
	}

	api := operations.NewPromqlServerAPI(swaggerSpec)
	api.GetTreeHandler = operations.GetTreeHandlerFunc(getTreeHandler)
	server := restapi.NewServer(api)
	defer server.Shutdown()

	parser := flags.NewParser(server, flags.Default)
	parser.ShortDescription = "Swagger Petstore"
	parser.LongDescription = "This is a sample server Petstore server.\n\n[Learn about Swagger](http://swagger.wordnik.com) or join the IRC channel '#swagger' on irc.freenode.net.\n\nFor this sample, you can use the api key 'special-key' to test the authorization filters\n"

	server.ConfigureFlags()
	for _, optsGroup := range api.CommandLineOptionsGroups {
		_, err := parser.AddGroup(optsGroup.ShortDescription, optsGroup.LongDescription, optsGroup.Options)
		if err != nil {
			log.Fatalln(err)
		}
	}

	if _, err := parser.Parse(); err != nil {
		code := 1
		if fe, ok := err.(*flags.Error); ok {
			if fe.Type == flags.ErrHelp {
				code = 0
			}
		}
		os.Exit(code)
	}

	if err := server.Serve(); err != nil {
		log.Fatalln(err)
	}
}

func getTreeHandler(params operations.GetTreeParams) middleware.Responder {
	query, err := promql.ParseExpr(params.Query)
	if err != nil {
		return operations.NewGetTreeUnprocessableEntity().WithPayload(err.Error())
	}

	node := queryToTree(query)

	return operations.NewGetTreeOK().WithPayload(node)
}

func queryToTree(node promql.Node) *models.Node {
	returnN := &models.Node{
		Children: []*models.Node{},
	}
	if node == nil {
		return returnN
	}

	returnN.T = strings.Split(fmt.Sprintf("%T", node), ".")[1]
	returnN.V = node.String()

	switch n := node.(type) {
	case *promql.EvalStmt:
		returnN.Children = append(returnN.Children, queryToTree(n.Expr))

	case promql.Expressions:
		for _, e := range n {
			returnN.Children = append(returnN.Children, queryToTree(e))
		}
	case *promql.AggregateExpr:
		returnN.Children = append(returnN.Children, queryToTree(n.Expr))

	case *promql.BinaryExpr:
		returnN.Children = append(returnN.Children, queryToTree(n.LHS))
		returnN.Children = append(returnN.Children, queryToTree(n.RHS))

	case *promql.Call:
		returnN.Children = append(returnN.Children, queryToTree(n.Args))

	case *promql.ParenExpr:
		returnN.Children = append(returnN.Children, queryToTree(n.Expr))

	case *promql.UnaryExpr:
		returnN.Children = append(returnN.Children, queryToTree(n.Expr))

	case *promql.SubqueryExpr:
		returnN.Children = append(returnN.Children, queryToTree(n.Expr))

	case *promql.MatrixSelector, *promql.NumberLiteral, *promql.StringLiteral, *promql.VectorSelector:
		// nothing to do

	default:
		panic("promql.Tree: not all node types covered")
	}

	return returnN
}
