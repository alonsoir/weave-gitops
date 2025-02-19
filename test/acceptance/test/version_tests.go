package acceptance

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Weave GitOps Version Test", func() {

	var session *gexec.Session

	BeforeEach(func() {

		By("Given I have a gitops binary installed on my local machine", func() {
			Expect(FileExists(WEGO_BIN_PATH)).To(BeTrue())
		})
	})

	It("SmokeTest - Verify that command gitops version prints the version information", func() {

		By("When I run the command 'gitops version'", func() {
			session = runCommandAndReturnSessionOutput(WEGO_BIN_PATH + " version")
		})

		By("Then I should see the gitops version printed in format vm.n.n with newline character", func() {
			Eventually(session).Should(gbytes.Say("Current Version: v[0-3].[0-9].[0-9]\n"))
		})

		By("And git commit with commit id", func() {
			Eventually(session).Should(gbytes.Say("GitCommit: ([a-f0-9]{7})|([a-f0-9]{8}\n)"))
		})

		By("And build timestamp", func() {
			Eventually(session).Should(gbytes.Say("BuildTime: [0-9-_:]+\n"))
		})

		By("And flux version", func() {
			Eventually(session).Should(gbytes.Say("Flux Version: v[0-9].[0-9][0-9].[0-9]\n"))
		})
	})
})
