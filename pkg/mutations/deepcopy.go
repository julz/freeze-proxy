package mutations

import "k8s.io/apimachinery/pkg/runtime"

func (in *Pod) DeepCopyInto(out *Pod) {
	*out = *in
	in.Pod.DeepCopyInto(&out.Pod)
	return
}

func (in *Pod) DeepCopy() *Pod {
	if in == nil {
		return nil
	}
	out := new(Pod)
	in.DeepCopyInto(out)
	return out
}

func (in *Pod) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}
