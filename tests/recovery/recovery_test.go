package cos_test

import (
	"github.com/rancher-sandbox/cOS/tests/sut"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("cOS Recovery upgrade tests", func() {
	var s *sut.SUT

	BeforeEach(func() {
		s = sut.NewSUT()
		s.EventuallyConnects()
	})

	Context("upgrading COS_ACTIVE from the recovery partition", func() {
		AfterEach(func() {
			if CurrentGinkgoTestDescription().Failed == false {
				s.Reset()
			}
		})
		It("upgrades to the latest", func() {
			currentName := s.GetOSRelease("NAME")

			By("booting into recovery to check the OS version")
			err := s.ChangeBootOnce(sut.Recovery)
			Expect(err).ToNot(HaveOccurred())

			s.Reboot()
			ExpectWithOffset(1, s.BootFrom()).To(Equal(sut.Recovery))

			recoveryName := s.GetOSRelease("NAME")

			// In these tests, if we are booting into squashfs we are booting into recovery. And the recovery image
			// is shipping a different os-release name (cOS recovery) instead of the standard one (cOS)
			if s.SquashFSRecovery() {
				Expect(currentName).ToNot(Equal(recoveryName))
			} else {
				Expect(currentName).To(Equal(recoveryName))
			}

			By("upgrade with CURRENT=active.img")
			out, err := s.Command("CURRENT=active.img cos-upgrade")
			Expect(err).ToNot(HaveOccurred())
			Expect(out).Should(ContainSubstring("Upgrade done, now you might want to reboot"))
			Expect(out).Should(ContainSubstring("Upgrading system"))

			By("Reboot to upgraded active")
			s.Reboot()
			ExpectWithOffset(1, s.BootFrom()).To(Equal(sut.Active))
		})

		It("upgrades to a specific image", func() {
			err := s.ChangeBoot(sut.Active)
			Expect(err).ToNot(HaveOccurred())

			s.Reboot()
			ExpectWithOffset(1, s.BootFrom()).To(Equal(sut.Active))
			currentVersion := s.GetOSRelease("VERSION")

			By("booting into recovery to check the OS version")
			err = s.ChangeBoot(sut.Recovery)
			Expect(err).ToNot(HaveOccurred())

			s.Reboot()
			ExpectWithOffset(1, s.BootFrom()).To(Equal(sut.Recovery))

			out, err := s.Command("CURRENT=active.img cos-upgrade --no-verify --docker-image quay.io/costoolkit/releases-opensuse:cos-system-0.5.7")
			Expect(err).ToNot(HaveOccurred())
			Expect(out).Should(ContainSubstring("Upgrade done, now you might want to reboot"))

			err = s.ChangeBoot(sut.Active)
			Expect(err).ToNot(HaveOccurred())

			s.Reboot()
			ExpectWithOffset(1, s.BootFrom()).To(Equal(sut.Active))

			upgradedVersion := s.GetOSRelease("VERSION")
			Expect(upgradedVersion).ToNot(Equal(currentVersion))
			Expect(upgradedVersion).To(Equal("0.5.7\n"))
		})
	})

	// After this tests the VM is no longer in its initial state!!
	Context("upgrading recovery", func() {
		When("using specific images", func() {
			It("upgrades to a specific image and reset back to the installed version", func() {
				version := s.GetOSRelease("VERSION")
				By("upgrading to quay.io/costoolkit/test-images:squashfs-recovery-image (0.5.3..) ")
				out, err := s.Command("cos-upgrade --no-verify --recovery --docker-image quay.io/costoolkit/test-images:squashfs-recovery-image")
				Expect(err).ToNot(HaveOccurred())
				Expect(out).Should(ContainSubstring("Upgrade done, now you might want to reboot"))
				Expect(out).Should(ContainSubstring("Upgrading recovery partition"))

				By("booting into recovery to check the OS version")
				err = s.ChangeBootOnce(sut.Recovery)
				Expect(err).ToNot(HaveOccurred())

				s.Reboot()
				ExpectWithOffset(1, s.BootFrom()).To(Equal(sut.Recovery))

				out = s.GetOSRelease("VERSION")
				Expect(out).ToNot(Equal(""))
				Expect(out).ToNot(Equal(version))
				Expect(out).To(Equal("0.5.3\n"))

				By("rebooting back to active")
				s.Reboot()
				ExpectWithOffset(1, s.BootFrom()).To(Equal(sut.Active))
			})
		})

		When("using upgrade channel", func() {
			It("upgrades to latest image", func() {
				By("upgrading recovery")
				out, err := s.Command("cos-upgrade --no-verify --recovery")
				Expect(err).ToNot(HaveOccurred())
				Expect(out).Should(ContainSubstring("Upgrade done, now you might want to reboot"))
				Expect(out).Should(ContainSubstring("Upgrading recovery partition"))

				By("Reboot to upgraded recovery")
				err = s.ChangeBootOnce(sut.Recovery)
				Expect(err).ToNot(HaveOccurred())
				s.Reboot()
				ExpectWithOffset(1, s.BootFrom()).To(Equal(sut.Recovery))
				By("rebooting back to active")
				s.Reboot()
				ExpectWithOffset(1, s.BootFrom()).To(Equal(sut.Active))
			})
		})

	})
})
