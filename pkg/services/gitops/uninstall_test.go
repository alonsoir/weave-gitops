package gitops_test

import (
	"context"
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/manifests"
	"github.com/weaveworks/weave-gitops/pkg/flux/fluxfakes"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"
	"github.com/weaveworks/weave-gitops/pkg/logger/loggerfakes"
	"github.com/weaveworks/weave-gitops/pkg/services/gitops"
)

var uninstallParams gitops.UninstallParams

func checkFluxUninstallFailure() {
	fluxErrMsg := "flux uninstall failed"

	loggedMsg := ""
	logger.PrintfStub = func(str string, args ...interface{}) {
		loggedMsg = fmt.Sprintf(str, args...)
	}

	fluxClient.UninstallStub = func(namespace string, dryRun bool) error {
		return errors.New(fluxErrMsg)
	}

	err := gitopsSrv.Uninstall(uninstallParams)

	Expect(loggedMsg).To(Equal(fmt.Sprintf("received error uninstalling flux: %q, continuing with uninstall", fluxErrMsg)))
	Expect(err).To(MatchError(gitops.UninstallError{}))
	Expect(kubeClient.GetClusterStatusCallCount()).To(Equal(1))
	Expect(fluxClient.UninstallCallCount()).To(Equal(1))
	namespace, dryRun := fluxClient.UninstallArgsForCall(0)
	Expect(namespace).To(Equal(wego.DefaultNamespace))
	Expect(dryRun).To(Equal(false))
}

func checkAppCRDUninstallFailure() {
	manifestsErrMsg := "gitops manifests uninstall failed"

	loggedMsg := ""
	logger.PrintfStub = func(str string, args ...interface{}) {
		loggedMsg = fmt.Sprintf(str, args...)
	}

	kubeClient.DeleteStub = func(ctx context.Context, manifest []byte) error {
		return errors.New(manifestsErrMsg)
	}

	err := gitopsSrv.Uninstall(uninstallParams)

	Expect(loggedMsg).To(Equal(fmt.Sprintf("received error deleting App CRD: %q", manifestsErrMsg)))
	Expect(err).To(MatchError(gitops.UninstallError{}))
	Expect(kubeClient.GetClusterStatusCallCount()).To(Equal(1))
	Expect(fluxClient.UninstallCallCount()).To(Equal(1))
	Expect(kubeClient.DeleteCallCount()).To(Equal(1))

	namespace, dryRun := fluxClient.UninstallArgsForCall(0)
	Expect(namespace).To(Equal(wego.DefaultNamespace))
	Expect(dryRun).To(Equal(false))
}

var _ = Describe("Uninstall", func() {
	BeforeEach(func() {
		fluxClient = &fluxfakes.FakeFlux{}
		kubeClient = &kubefakes.FakeKube{}
		logger = &loggerfakes.FakeLogger{}
		gitopsSrv = gitops.New(logger, fluxClient, kubeClient)

		uninstallParams = gitops.UninstallParams{
			Namespace: wego.DefaultNamespace,
			DryRun:    false,
		}
	})

	It("logs warning information if wego is not installed before proceeding", func() {
		err := gitopsSrv.Uninstall(uninstallParams)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(kubeClient.GetClusterStatusCallCount()).To(Equal(1))
		Expect(fluxClient.UninstallCallCount()).To(Equal(1))

		loggedMsg := ""
		logger.PrintlnStub = func(str string, args ...interface{}) {
			loggedMsg = str
		}

		kubeClient.GetClusterStatusStub = func(ctx context.Context) kube.ClusterStatus {
			return kube.FluxInstalled
		}

		Expect(gitopsSrv.Uninstall(uninstallParams)).Should(Succeed())
		Expect(loggedMsg).To(Equal("gitops is not fully installed... removing any partial installation\n"))

		kubeClient.GetClusterStatusStub = func(ctx context.Context) kube.ClusterStatus {
			return kube.Unmodified
		}
		loggedMsg = ""

		Expect(gitopsSrv.Uninstall(uninstallParams)).Should(Succeed())
		Expect(loggedMsg).To(Equal("gitops is not fully installed... removing any partial installation\n"))
	})

	It("Does not log warning information if wego is installed", func() {
		kubeClient.GetClusterStatusStub = func(ctx context.Context) kube.ClusterStatus {
			return kube.GitOpsInstalled
		}

		loggedMsg := ""
		logger.PrintlnStub = func(str string, args ...interface{}) {
			loggedMsg = str
		}

		Expect(gitopsSrv.Uninstall(uninstallParams)).Should(Succeed())
		Expect(loggedMsg).To(Equal(""))
	})

	It("Generates an error if flux uninstall fails with wego installed", func() {
		kubeClient.GetClusterStatusStub = func(ctx context.Context) kube.ClusterStatus {
			return kube.GitOpsInstalled
		}

		checkFluxUninstallFailure()
	})

	It("Generates an error if flux uninstall fails with only flux installed", func() {
		kubeClient.GetClusterStatusStub = func(ctx context.Context) kube.ClusterStatus {
			return kube.FluxInstalled
		}

		checkFluxUninstallFailure()
	})

	It("Generates an error if flux uninstall fails with partial or no flux installed", func() {
		kubeClient.GetClusterStatusStub = func(ctx context.Context) kube.ClusterStatus {
			return kube.Unmodified
		}

		checkFluxUninstallFailure()
	})

	It("Generates an error if CRD uninstall fails with wego installed", func() {
		kubeClient.GetClusterStatusStub = func(ctx context.Context) kube.ClusterStatus {
			return kube.GitOpsInstalled
		}

		checkAppCRDUninstallFailure()
	})

	It("Generates an error if CRD uninstall fails with only flux installed", func() {
		kubeClient.GetClusterStatusStub = func(ctx context.Context) kube.ClusterStatus {
			return kube.FluxInstalled
		}

		checkAppCRDUninstallFailure()
	})

	It("Generates an error if CRD uninstall fails with partial or no flux installed", func() {
		kubeClient.GetClusterStatusStub = func(ctx context.Context) kube.ClusterStatus {
			return kube.Unmodified
		}

		checkAppCRDUninstallFailure()
	})

	It("deletes weave gitops manifests", func() {
		err := gitopsSrv.Uninstall(uninstallParams)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(kubeClient.DeleteCallCount()).To(Equal(1))

		_, appCRD := kubeClient.DeleteArgsForCall(0)
		Expect(appCRD).To(Equal(manifests.AppCRD))
	})

	Context("when dry-run", func() {
		BeforeEach(func() {
			uninstallParams.DryRun = true
		})

		It("calls flux uninstall", func() {
			err := gitopsSrv.Uninstall(uninstallParams)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(fluxClient.UninstallCallCount()).To(Equal(1))

			namespace, dryRun := fluxClient.UninstallArgsForCall(0)
			Expect(namespace).To(Equal(wego.DefaultNamespace))
			Expect(dryRun).To(Equal(true))
		})

		It("does not call kube apply", func() {
			err := gitopsSrv.Uninstall(uninstallParams)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(kubeClient.DeleteCallCount()).To(Equal(0))
		})
	})
})
