# -*- mode: ruby -*-
# vi: set ft=ruby :

# Vagrantfile API/syntax version. Don't touch unless you know what you're doing!
VAGRANTFILE_API_VERSION = "2"

Vagrant.require_version ">= 1.5.0"

$install_reqs = <<SCRIPT
  apt update
  apt install -y screen git python-pip golang
  pip install flask
  cd /home/vagrant
  ln -s /vagrant/httpd  ||:
  ln -s /vagrant lbapp  ||:
SCRIPT

$build_lb = <<SCRIPT
  cd lbapp
  # mkdir -pv ./go/bin
  # export GOPATH=${PWD}/go
  # export PATH=$PATH:${GOPATH}/bin
  go build lb.go
SCRIPT


Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|

  config.vm.define "ubuntu" do |ubuntu|
    ubuntu.vm.box = "ubuntu/xenial64"
    ubuntu.vm.box_check_update = false
    ubuntu.vm.hostname = "lbvm"
    ubuntu.vm.network :forwarded_port, guest: 8080, host: 8080

    ubuntu.vm.provider :virtualbox do |vb|
      vb.cpus = 2
      vb.customize ["modifyvm", :id, "--memory", "256"]
      vb.customize ["modifyvm", :id, "--nictype1", "virtio"]
    end

    ubuntu.vm.provision "install_reqs", type: "shell", inline: $install_reqs
    ubuntu.vm.provision "build_lb", type: "shell", inline: $build_lb, privileged: false

  end
end
