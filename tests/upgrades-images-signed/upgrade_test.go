package cos_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/cOS/tests/sut"
)

var _ = Describe("cOS Upgrade tests - Images signed", func() {
	var s *sut.SUT

	BeforeEach(func() {
		s = sut.NewSUT()
		s.EventuallyConnects(360)
	})

	AfterEach(func() {
		s.Reset()
	})
	Context("After install", func() {
		When("images are signed", func() {
			It("upgrades to a specific image and reset back to the installed version", func() {
				out, err := s.Command("source /etc/os-release && echo $VERSION")
				Expect(err).ToNot(HaveOccurred())
				Expect(out).ToNot(Equal(""))

				version := out
				By("upgrading to an old signed image")
				out, err = s.Command("cos-upgrade --verify --docker-image raccos/releases-opensuse:cos-system-0.4.32")
				Expect(err).ToNot(HaveOccurred())
				Expect(out).Should(ContainSubstring("Upgrade done, now you might want to reboot"))
				Expect(out).Should(ContainSubstring("Booting from: active.img"))

				By("rebooting and checking out the version")
				s.Reboot()

				out, err = s.Command("source /etc/os-release && echo $VERSION")
				Expect(err).ToNot(HaveOccurred())
				Expect(out).ToNot(Equal(""))
				Expect(out).ToNot(Equal(version))
				Expect(out).To(Equal("0.4.32\n"))
			})
			// TODO(itxaka): Is this basically the same test as the one below??
			It("fails to upgrade if verify is enabled on an unsigned version", func() {
				// Using releases-amd64 as those images are not signed
				out, err := s.Command("cos-upgrade --verify --docker-image raccos/releases-amd64:cos-system-0.4.18")
				Expect(err).To(HaveOccurred())
				Expect(out).Should(ContainSubstring("No valid trust data"))
			})
			It("fails to upgrade if verify is enabled on an unsigned upgrade channel", func() {
				out, err := s.Command("sed -i 's|raccos/releases-.*|raccos/releases-amd64\"|g' /etc/luet/luet.yaml && cos-upgrade --verify")
				Expect(out).Should(ContainSubstring("does not have trust data"))
				Expect(err).To(HaveOccurred())
			})
			It("upgrades to an signed image with --verify and can reset back to the installed state", func() {
				out, err := s.Command("source /etc/os-release && echo $VERSION")
				Expect(err).ToNot(HaveOccurred())
				Expect(out).ToNot(Equal(""))

				version := out

				By("running cos-upgrade with --verify and an signed image")
				out, err = s.Command("cos-upgrade --verify --docker-image raccos/releases-opensuse:cos-system-0.4.9-9")
				Expect(err).ToNot(HaveOccurred())
				Expect(out).Should(ContainSubstring("Upgrade done, now you might want to reboot"))
				Expect(out).Should(ContainSubstring("to /usr/local/tmp/rootfs"))
				Expect(out).Should(ContainSubstring("Booting from: active.img"))

				By("rebooting and checking out the version")
				s.Reboot()

				out, err = s.Command("source /etc/os-release && echo $VERSION")
				Expect(err).ToNot(HaveOccurred())
				Expect(out).ToNot(Equal(""))
				Expect(out).ToNot(Equal(version))
				Expect(out).To(Equal("0.4.31\n"))

				By("rollbacking state")
				s.Reset()

				out, err = s.Command("source /etc/os-release && echo $VERSION")
				Expect(err).ToNot(HaveOccurred())
				Expect(out).ToNot(Equal(""))
				Expect(out).ToNot(Equal("0.4.9-9\n"))
				Expect(out).To(Equal(version))
			})
		})
	})
})
