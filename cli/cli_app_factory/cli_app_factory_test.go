package cli_app_factory_test

import (
	"errors"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/codegangsta/cli"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-cf-experimental/lattice-cli/cli/cli_app_factory"
	"github.com/pivotal-cf-experimental/lattice-cli/cli/config"
	"github.com/pivotal-cf-experimental/lattice-cli/cli/config/persister"
	"github.com/pivotal-cf-experimental/lattice-cli/cli/config/target_verifier/fake_target_verifier"
	"github.com/pivotal-cf-experimental/lattice-cli/cli/exit_handler/fake_exit_handler"
	"github.com/pivotal-cf-experimental/lattice-cli/cli/output"
	"github.com/pivotal-cf-experimental/lattice-cli/cli/test_helpers"
	"github.com/pivotal-golang/lager"

	config_command_factory "github.com/pivotal-cf-experimental/lattice-cli/cli/config/command_factory"
)

var _ = Describe("CliAppFactory", func() {
	var (
		fakeTargetVerifier *fake_target_verifier.FakeTargetVerifier
		memPersister       persister.Persister
		outputBuffer       *gbytes.Buffer
		cliApp             *cli.App
		cliConfig          *config.Config
	)
	BeforeEach(func() {
		fakeTargetVerifier = &fake_target_verifier.FakeTargetVerifier{}
		memPersister = persister.NewMemPersister()
		outputBuffer = gbytes.NewBuffer()
		cliConfig = config.New(memPersister)
		cliApp = cli_app_factory.MakeCliApp(
			"30",
			"~/",
			&fake_exit_handler.FakeExitHandler{},
			cliConfig,
			lager.NewLogger("test"),
			fakeTargetVerifier,
			output.New(outputBuffer),
		)
	})

	Describe("MakeCliApp", func() {
		It("makes an app", func() {
			Expect(cliApp).ToNot(BeNil())
			Expect(cliApp.Name).To(Equal("ltc"))
			Expect(cliApp.Usage).To(Equal(cli_app_factory.LtcUsage))
			Expect(cliApp.Commands).NotTo(BeEmpty())
		})

		Describe("App's before Action", func() {
			Context("when running the target command", func() {
				It("does not verify the current target", func() {
					cliConfig.SetTarget("my-lattice.example.com")
					cliConfig.Save()

					commandRan := false

					cliApp.Commands = []cli.Command{
						cli.Command{
							Name: config_command_factory.TargetCommandName,
							Action: func(ctx *cli.Context) {
								commandRan = true
							},
						},
					}

					cliAppArgs := []string{"ltc", config_command_factory.TargetCommandName}

					err := cliApp.Run(cliAppArgs)

					Expect(err).ToNot(HaveOccurred())
					Expect(fakeTargetVerifier.VerifyTargetCallCount()).To(Equal(0))
					Expect(commandRan).To(Equal(true))
				})
			})

			Context("when running the help command", func() {
				It("does not verify the current target", func() {
					cliConfig.SetTarget("my-lattice.example.com")
					cliConfig.Save()

					commandRan := false

					cliApp.Commands = []cli.Command{
						cli.Command{
							Name: "help",
							Action: func(ctx *cli.Context) {
								commandRan = true
							},
						},
					}

					cliAppArgs := []string{"ltc", "help"}

					err := cliApp.Run(cliAppArgs)

					Expect(err).ToNot(HaveOccurred())
					Expect(fakeTargetVerifier.VerifyTargetCallCount()).To(Equal(0))
					Expect(commandRan).To(Equal(true))
				})
			})
			Context("when running the bare ltc command", func() {
				It("does not verify the current target", func() {
					cliConfig.SetTarget("my-lattice.example.com")
					cliConfig.Save()

					commandRan := false
					cliApp.Action = func(context *cli.Context) {
						commandRan = true
					}

					cliAppArgs := []string{"ltc"}

					err := cliApp.Run(cliAppArgs)

					Expect(err).ToNot(HaveOccurred())
					Expect(fakeTargetVerifier.VerifyTargetCallCount()).To(Equal(0))
				})
			})

			Context("when we cannot find the subcommand", func() {
				It("does not verify the current target", func() {
					cliConfig.SetTarget("my-lattice.example.com")
					cliConfig.Save()

					commandRan := false
					cliApp.Action = func(context *cli.Context) {
						commandRan = true
					}

					cliAppArgs := []string{"ltc", "buy-me-a-pony"}

					err := cliApp.Run(cliAppArgs)

					Expect(err).ToNot(HaveOccurred())
					Expect(fakeTargetVerifier.VerifyTargetCallCount()).To(Equal(0))
				})
			})

			Context("Any other command", func() {
				Context("when targeted receptor is up and we are authorized", func() {
					It("executes the command", func() {
						fakeTargetVerifier.VerifyTargetReturns(true, true, nil)

						cliConfig.SetTarget("my-lattice.example.com")
						cliConfig.Save()

						commandRan := false
						cliApp.Commands = []cli.Command{cli.Command{Name: "print-a-unicorn", Action: func(ctx *cli.Context) {
							commandRan = true
						}}}

						cliAppArgs := []string{"ltc", "print-a-unicorn"}

						err := cliApp.Run(cliAppArgs)

						Expect(err).ToNot(HaveOccurred())
						Expect(fakeTargetVerifier.VerifyTargetCallCount()).To(Equal(1))
						Expect(fakeTargetVerifier.VerifyTargetArgsForCall(0)).To(Equal("http://receptor.my-lattice.example.com"))
						Expect(commandRan).To(BeTrue())

					})
				})

				Context("when we are unauthorized for the targeted receptor", func() {
					It("Prints an error message and does not execute the command", func() {
						fakeTargetVerifier.VerifyTargetReturns(true, false, nil)
						cliConfig.SetTarget("my-borked-lattice.example.com")
						cliConfig.Save()

						commandRan := false
						cliApp.Commands = []cli.Command{cli.Command{Name: "print-a-unicorn", Action: func(ctx *cli.Context) { commandRan = true }}}

						cliAppArgs := []string{"ltc", "print-a-unicorn"}

						err := cliApp.Run(cliAppArgs)

						Expect(err).To(HaveOccurred())
						Expect(outputBuffer).To(test_helpers.Say("Could not authenticate with the receptor. Please run ltc target with the correct credentials."))
						Expect(fakeTargetVerifier.VerifyTargetCallCount()).To(Equal(1))
						Expect(fakeTargetVerifier.VerifyTargetArgsForCall(0)).To(Equal("http://receptor.my-borked-lattice.example.com"))
						Expect(commandRan).To(BeFalse())
					})
				})

				Context("when the receptor is down", func() {
					It("prints a helpful error", func() {
						fakeTargetVerifier.VerifyTargetReturns(false, false, errors.New("oopsie!"))

						cliConfig.SetTarget("my-borked-lattice.example.com")
						cliConfig.Save()

						commandRan := false
						cliApp.Commands = []cli.Command{cli.Command{Name: "print-a-unicorn", Action: func(ctx *cli.Context) { commandRan = true }}}

						cliAppArgs := []string{"ltc", "print-a-unicorn"}

						err := cliApp.Run(cliAppArgs)

						Expect(err).To(HaveOccurred())
						Expect(outputBuffer).To(test_helpers.Say("Error connecting to the receptor. Make sure your lattice target is set, and that lattice is up and running.\n\tUnderlying error: oopsie!"))
						Expect(fakeTargetVerifier.VerifyTargetCallCount()).To(Equal(1))
						Expect(fakeTargetVerifier.VerifyTargetArgsForCall(0)).To(Equal("http://receptor.my-borked-lattice.example.com"))
						Expect(commandRan).To(BeFalse())
					})
				})
			})
		})

	})

	Describe("Timeout", func() {
		It("returns the timeout in seconds", func() {
			Expect(cli_app_factory.Timeout("25")).To(Equal(25 * time.Second))
		})

		It("returns one minute for an empty string", func() {
			Expect(cli_app_factory.Timeout("")).To(Equal(time.Minute))
		})

		It("returns one minute for an invalid string", func() {
			Expect(cli_app_factory.Timeout("CANNOT PARSE")).To(Equal(time.Minute))
		})
	})

	Describe("LoggregatorUrl", func() {
		It("returns loggregator url with the websocket scheme added", func() {
			loggregatorUrl := cli_app_factory.LoggregatorUrl("doppler.diego.io")
			Expect(loggregatorUrl).To(Equal("ws://doppler.diego.io"))
		})
	})

})
