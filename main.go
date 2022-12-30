package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/ssm"
)

// Cluster represents a kubernetes cluster
type Cluster struct {
	Arn, Endpoint, Env, Name, CertificateAuthority      string
	OidcClientIssuerUrl, OidcClientID, OidcClientSecret string
}

type arrayFlags []string

var (
	clusterFlags     arrayFlags
	printVersionFlag bool
	BuildVersion     string = "development"
)

func (i *arrayFlags) String() string {
	return fmt.Sprintf("[%s]", strings.Join(*i, " "))
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func init() {
	const (
		defaultClusterFlags      = "dev:dev1"
		defaultClusterFlagsUsage = "AWS Profile and EKS Cluster name(s) joined by a colon, can be passed more than once\ne.g. -clusters dev:dev1 -clusters tst:tst1"
		defaultVersionFlag       = false
		defaultVersionFlagUsage  = "print current version and exit"
	)
	flag.BoolVar(&printVersionFlag, "version", defaultVersionFlag, defaultVersionFlagUsage)
	flag.BoolVar(&printVersionFlag, "v", defaultVersionFlag, defaultVersionFlagUsage+" (shorthand)")
	flag.Var(&clusterFlags, "clusters", defaultClusterFlagsUsage)
}

func main() {
	flag.Parse()
	if printVersionFlag {
		fmt.Fprintf(os.Stderr, "%s\n", BuildVersion)
		os.Exit(0)
	}
	clusters := []Cluster{}
	for _, c := range clusterFlags {
		cInfo := strings.Split(c, ":")
		// if len(cInfo) < 2 ERROR!
		fmt.Fprintf(os.Stderr, "Describing cluster %s in env/profile %s\n", cInfo[1], cInfo[0])
		cluster := &Cluster{}
		err := getClusterInfo(cluster, cInfo[1], cInfo[0])
		if err != nil {
			panic(err)
		}
		cluster.OidcClientSecret, err = getOidcSecret(fmt.Sprintf("/okta/oidc/%s/secret", cInfo[1]), cInfo[0])
		if err != nil {
			panic(err)
		}
		clusters = append(clusters, *cluster)
	}
	kubeConfigBuf := &bytes.Buffer{}
	err := genKubeConfig(clusters, kubeConfigBuf)
	if err != nil {
		panic(nil)
	}
	fmt.Printf("\n%s\n", kubeConfigBuf)
}

func getOidcSecret(secretPath, profile string) (string, error) {
	// AWS SSM Client/Session Setup
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile:           profile,
	})
	if err != nil {
		return "", err
	}
	svc := ssm.New(sess)

	getParameterInput := &ssm.GetParameterInput{
		Name:           aws.String(secretPath),
		WithDecryption: aws.Bool(true),
	}
	OidcParameterInfo, err := svc.GetParameter(getParameterInput)
	if err != nil {
		return "", err
	}
	return aws.StringValue(OidcParameterInfo.Parameter.Value), nil
}

func getClusterInfo(cluster *Cluster, clusterName, profile string) error {
	// AWS EKS Client/Session Setup
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile:           profile,
	})
	if err != nil {
		panic(err)
	}
	svc := eks.New(sess)

	// Gather Cluster, IdentityProvider information
	describeClusterInput := &eks.DescribeClusterInput{
		Name: aws.String(clusterName),
	}
	clusterInfo, err := svc.DescribeCluster(describeClusterInput)
	if err != nil {
		panic(err)
	}
	listIdentityProviderConfigsInput := &eks.ListIdentityProviderConfigsInput{
		ClusterName: aws.String(clusterName),
	}
	identityProviderConfigsInfo, err := svc.ListIdentityProviderConfigs(listIdentityProviderConfigsInput)
	if err != nil {
		panic(err)
	}
	describeIdentityProviderConfigInput := &eks.DescribeIdentityProviderConfigInput{
		ClusterName:            aws.String(clusterName),
		IdentityProviderConfig: identityProviderConfigsInfo.IdentityProviderConfigs[0],
	}
	identityProviderConfigDetails, err := svc.DescribeIdentityProviderConfig(describeIdentityProviderConfigInput)
	if err != nil {
		panic(err)
	}

	cluster.Arn = aws.StringValue(clusterInfo.Cluster.Arn)
	cluster.Env = profile
	cluster.Name = clusterName
	cluster.Endpoint = aws.StringValue(clusterInfo.Cluster.Endpoint)
	cluster.OidcClientIssuerUrl = aws.StringValue(identityProviderConfigDetails.IdentityProviderConfig.Oidc.IssuerUrl)
	cluster.OidcClientID = aws.StringValue(identityProviderConfigDetails.IdentityProviderConfig.Oidc.ClientId)
	cluster.OidcClientSecret = ""
	cluster.CertificateAuthority = aws.StringValue(clusterInfo.Cluster.CertificateAuthority.Data)

	return nil
}

func genKubeConfig(clusters []Cluster, kubeConfigBuf *bytes.Buffer) error {
	kubeConfigTmplData, err := os.ReadFile("config.tmpl")
	if err != nil {
		return err
	}
	kubeConfigTmpl, err := template.New("config").Parse(string(kubeConfigTmplData))
	if err != nil {
		return err
	}

	err = kubeConfigTmpl.Execute(kubeConfigBuf, clusters)
	if err != nil {
		return err
	}
	return nil
}
