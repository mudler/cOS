package cos_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/cOS/tests/sut"
)

var _ = Describe("cOS Deploy tests", func() {
	var s *sut.SUT
	var isVagrant bool

	BeforeSuite(func() {
		isVagrant = sut.IsVagrantTest()
		if isVagrant {
			sut.SnapshotVagrant()
		}
	})

	AfterSuite(func() {
		if isVagrant {
			sut.SnapshotVagrantDelete()
		}
	})

	BeforeEach(func() {
		s = sut.NewSUT()
		s.EventuallyConnects(360)
	})

	AfterEach(func() {
		// Try to gather mtree logs on failure
		if CurrentGinkgoTestDescription().Failed {
			s.GatherLog("/tmp/image-mtree-check.log")
			s.GatherLog("/tmp/luet_mtree_failures.log")
			s.GatherLog("/tmp/luet_mtree.log")
			s.GatherLog("/tmp/luet.log")
		}
		if CurrentGinkgoTestDescription().Failed == false {
			if isVagrant {
				sut.ResetWithVagrant()
			} else {
				s.Reset()
			}

		}
	})
	Context("After install", func() {
		When("deploying again", func() {
			It("deploys only if --force flag is provided", func() {
				By("deploying without --force")
				out, err := s.Command("cos-deploy --docker-image quay.io/costoolkit/releases-opensuse:cos-system-0.5.5")
				Expect(out).Should(ContainSubstring("There is already an active deployment"))
				Expect(err).To(HaveOccurred())
				By("deploying with --force")
				out, err = s.Command("cos-deploy --force --docker-image quay.io/costoolkit/releases-opensuse:cos-system-0.5.5")
				Expect(out).Should(ContainSubstring("Forcing overwrite"))
				Expect(out).Should(ContainSubstring("now you might want to reboot"))
				Expect(err).NotTo(HaveOccurred())
			})
			It("force deploys from recovery", func() {
				err := s.ChangeBoot(sut.Recovery)
				Expect(err).ToNot(HaveOccurred())
				s.Reboot()
				ExpectWithOffset(1, s.BootFrom()).To(Equal(sut.Recovery))
				By("deploying with --force")
				out, err := s.Command("cos-deploy --force --docker-image quay.io/costoolkit/releases-opensuse:cos-system-0.5.5")
				Expect(out).Should(ContainSubstring("now you might want to reboot"))
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
