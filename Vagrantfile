provision = %(
# install golang 1.5 so we can use it to bootstrap go1.6 next
echo "deb http://ftp.debian.org/debian jessie-backports main" > backports.list
sudo mv backports.list /etc/apt/sources.list.d/
sudo apt-get update
sudo apt-get install -t jessie-backports -y golang

# needed for journalbeat development
sudo apt-get install -y libsystemd-dev

# install go1.6 from source since we don't have packages for jessie
wget https://github.com/golang/go/archive/go1.6.2.tar.gz
sudo tar -xzf go1.6.2.tar.gz -C /usr/local/lib go-go1.6.2
sudo ln -s /usr/local/lib/go-go1.6.2/ /usr/local/lib/go1.6
cd /usr/local/lib/go1.6/src
sudo GOROOT_BOOTSTRAP=/usr/lib/go ./make.bash

# set up the go1.6 dev environment
mkdir -p /opt/go
cat << EOF > /home/vagrant/.bashrc
export PATH=/usr/local/lib/go1.6/bin:$PATH
export GOROOT=/usr/local/lib/go1.6
export GOPATH=/opt/go
EOF
)

Vagrant.configure('2') do |config|
  vm_ram = ENV['VAGRANT_VM_RAM'] || 1024
  vm_cpu = ENV['VAGRANT_VM_CPU'] || 1

  config.vm.box = "debian/jessie64"

  config.vm.provider :virtualbox do |vb|
    vb.customize ["modifyvm", :id, "--memory", vm_ram, "--cpus", vm_cpu]
  end

  config.vm.network :private_network, ip: "172.17.8.99"

  # mount in ~/.aws so AWS creds are available to the tool
  config.vm.synced_folder ENV['HOME'] + "/src/journalbeat/", "/opt/go/src/github.com/bcwaldon/journalbeat", nfs: true, mount_options: ['nolock,vers=3,udp']

  config.vm.provision :shell, :inline => provision, :privileged => true
end
