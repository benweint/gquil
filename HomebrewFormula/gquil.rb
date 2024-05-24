class Gquil < Formula
  version 'v0.1.3'
  desc "Inspect, visualize, and transform GraphQL schemas on the command line."
  homepage "https://github.com/benweint/gquil"

  if OS.mac?
    url "https://github.com/benweint/gquil/releases/download/#{version}/gquil_darwin_amd64"
    sha256 "fe721caaa56df93aed0c730d2c5e5fee06e34b3a2cfd3886e03810399169d722"
  elsif OS.linux?
    url "https://github.com/benweint/gquil/releases/download/#{version}/gquil_linux_amd64"
    sha256 "1d8b3eee8ed2dda3d6a25d849c724806df12ee521b558305194012c85195b44c"
  end

  def install
    if OS.mac?
      bin.install "gquil_darwin_amd64" => "gquil"
    elsif OS.linux?
      bin.install "gquil_linux_amd64" => "gquil"
    end
  end
end
