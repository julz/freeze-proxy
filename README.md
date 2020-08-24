# What is `pauseme``?

A knative add-on that uses cgroups to pause/unpause containers when request
count drops to zero.

_Status_: PoC - not for production. But fun/interesting!

![screencast of pauseme](demo/demo.gif)

# Why?

This is an experiment in enabling a lambda-style UX where containers are paused
between requests. This allows you to keep "warm" containers around than can be
quickly unpaused when load comes in, without needing to worry about people
using that background capacity without paying for it. Particularly useful in
combination with the `scale-down-delay` paramter to keep containers around for
(e.g.) 15 minutes after the request count drops, avoiding a cold start penalty
in this case.

tl;dr: if you're charging by request count/length, this stops apps using cpu
resources when not processing a request. Used in combination with more generous
scale down times, you can avoid cold start penalties for many workloads.

# How?

Pauseme runs a mutating admission controller that modifies knative pods to
introduce a small proxy server -- the freeze proxy -- between the queue proxy
and the user container.  If request count falls to zero the freeze proxy uses a
mounted containerd socket to pause the user's container. When a new request
comes in, the user's container is resumed in the same way, and the proxy
forwards the request over.

# Install

1. Ensure your Kubernetes is using `containerd` as the container runtime.
  - e.g. for minikube, add `--runtime containerd` to `minikube start` command.
  - Support for other container runtimes should be possible, eventually.
1. Install knative as normal
1. `ko apply -f config/webhook.yaml`
1. That's it - deploy your knative service as normal!

# Example App

The "sleeptalker" app (which can be deployed via `ko apply -f
config/example.yaml`) runs a background goroutine that prints a message every
couple of seconds while it is running. If you deploy it normally and watch the
logs you'll see it printing constantly. If you deploy it after installing
`pauseme`, you'll see it only manages to print while the server is servicing
http requests.

# Limitations / Future Work

Known limitations (there may be more, this is a PoC!) / Future Work:

 - Only works with containerd right now, though support for cri-o/docker
   shouldn't be impossibly hard.
 - Mounts the containerd socket in to the freeze container, which requires root
   to access, which means the freeze-proxy sidecar runs as root. This could be
   avoided by using an intermediate daemonset, though. 
 - Mounts the containerd socket in to the freeze container, which requires
   allowing hostPath volume mounts. User containers can't use this to do
   anything nasty since knative doesn't permit it, but this stops you having as
   secure a PSP as would be ideal.
 - Doesn't work with knative's multi-container featre flag yet (though it
   should probably work fine eventually).
 - Pauses immediately when request count hits zero, might be nice to wait a few
   milliseconds in case another request comes in to save the small overhead of
   the pause/unpause in this case.
 - Todo: Graceful shutdown isn't quite right, we exit immediately when we should keep
   proxying until shutdown is complete.
 - Todo: freeze should answer liveness/readiness probes when
   user-container is frozen rather than forwarding.
