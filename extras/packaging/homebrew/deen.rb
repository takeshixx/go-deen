class Deen < Formula
  desc "Encode, decode, hash, compress, and format data"
  homepage "https://deen.adversec.com"
  url "https://github.com/takeshixx/go-deen/archive/refs/tags/v3.4.0.tar.gz"
  sha256 "008f5844fb485d8279c44a7733cf5e32d61ac8ff43e3380ed2d8b79291126fb5"
  license "Apache-2.0"

  depends_on "go" => :build

  def install
    ldflags = %W[
      -s -w
      -X github.com/takeshixx/deen/internal/core.version=v#{version}
    ]

    ENV["CGO_ENABLED"] = "0"
    system "go", "build", "-mod=readonly", *std_go_args(ldflags:), "./cmd/deen"
  end

  test do
    assert_equal "v#{version}", shell_output("#{bin}/deen -version").strip
    assert_equal "dGVzdA==", pipe_output("#{bin}/deen base64", "test")
  end
end
