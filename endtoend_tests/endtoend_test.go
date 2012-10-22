package endtoend_test

import (
	"flag"
	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/ec2"
	. "launchpad.net/gocheck"
	"testing"
	"time"
)

func Test(t *testing.T) { TestingT(t) }

type S struct {
	vm *VM
}

var _ = Suite(&S{})

var flagDesc = "enable end-to-end tests that creates a machine in amazon, you'll need a AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY to run this tests."
var enableSuite = flag.Bool("endtoend", false, flagDesc)

type VM struct {
	instanceId string
	ec2        *ec2.EC2
}

func (s *S) stopOnStateChange(toState string, c *C) {
	ticker := time.Tick(time.Minute)
	for _ = range ticker {
		instResp, err := s.vm.ec2.Instances([]string{s.vm.instanceId}, nil)
		c.Check(err, IsNil)
		state := instResp.Reservations[0].Instances[0].State
		if state.Name == toState {
			break
		}
	}
}

func (s *S) newVM(c *C) {
	auth, err := aws.EnvAuth()
	c.Check(err, IsNil)
	e := ec2.New(auth, aws.USEast)
	s.vm = &VM{ec2: e}
	options := ec2.RunInstances{
		ImageId:      "ami-ccf405a5", // ubuntu maverik
		InstanceType: "t1.micro",
	}
	resp, err := e.RunInstances(&options)
	c.Check(err, IsNil)
	instanceId := resp.Instances[0].InstanceId
	s.vm.instanceId = instanceId
	// wait until instance is up
	s.stopOnStateChange("running", c)
}

func (s *S) destroyVM(c *C) {
	_, err := s.vm.ec2.TerminateInstances([]string{s.vm.instanceId})
	c.Check(err, IsNil)
	s.stopOnStateChange("terminated", c)
}

func (s *S) SetUpSuite(c *C) {
	if !*enableSuite {
		c.Skip("skipping end-to-end suite, use -endtoend to enable")
	}
	s.newVM(c)
}

func (s *S) TearDown(c *C) {
	s.destroyVM(c)
}

func (s *S) TestTrueIsTrue(c *C) {
	c.Assert(true, Equals, true)
}
