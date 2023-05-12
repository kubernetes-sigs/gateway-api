# GEP-X: Declarative Policy

* Issue: TODO
* Status: Provisional

## Definitions

In this document we'll use `Policy` to refer to things that are specifically called policies
as well as other "MetaResources" that follow similar patterns.

## TLDR

This proposal is a follow-up to [GEP-713 Metaresources and Policy Attachment][713] to recommend
that we specifically remove the "attachment" part of "policy attachment" in favor of something
that is declarative at the affected resource level.

[713]:https://gateway-api.sigs.k8s.io/geps/gep-713/

## Goals

- Remove "attachment" from `Policy` resources and related documentation.
- Retain `Policy` resource structure other than "attachment" semantics.
- Provide new semantics to incorporate `Policy` resources at the level of the `Resource` that
  will be affected.

## The Problem: A Parable of Jane

It's a sunny Wednesday afternoon, and the lead microservices developer for
Evil Genius Cupcakes is windsurfing. Work has been eating Jane alive for the
past two and a half weeks, but after successfully deploying version 3.6.0 of
the `baker` service that morning, she escaped early to try to unwind a bit.

Her shoulders are just starting to unknot when her phone pings with a text
from Julian, down in the NOC. Waterproof phones are a blessing, but also a
curse.

**Julian**: _Hey Jane. Things are still running, more or less, but latencies
on everything in the `baker` namespace are crazy high after your last rollout,
and `baker` itself has a weirdly high load. Sorry to interrupt you on the lake
but can you take a look? Thanks!!_

Jane stares at the phone for a long moment, heart sinking, then slowly tacks
back to shore to dry off and grab her laptop.

What she finds when she logs in is strange. `baker` is taking a _lot_ of load,
almost 4x what’s being reported by its usual clients, and its clients report
that calls are taking much longer than they’d expect them to. She doublechecks
the Deployment, the Service, and all the HTTPRoutes around `baker`; everything
looks good. `baker`’s logs show her mostly failed requests... with a lot of
duplicates? Jane checks her HTTPRoute again, though she's pretty sure you
can't configure retries there, and finds nothing. But it definitely looks like
a client is retrying when it shouldn’t be.

She pings Julian.

**Jane**: _Hey Julian. Something weird is up, looks like requests to `baker`
are failing but getting retried??_

A minute later he answers.

**Julian**: 🤷 _Did you configure retries?_

**Jane**: _Dude. I don’t even know how to._ 😂

**Julian**: _You just attach a RetryPolicy to your HTTPRoute._

**Jane**: _Nope. Definitely didn’t do that._

She types `kubectl get retrypolicy -n baker` and gets a permission error.

**Jane**: _Huh, I actually don’t have permissions for RetryPolicy._ 🤔

**Julian**: 🤷 _Feels like you should but OK, guess that can’t be it._

Minutes pass while both look at logs.

**Jane**: _OK, it’s definitely retrying. Nearly every request fails the first
few times, gets retried, and then finally succeeds?_

**Julian**: _Are you sure? I don’t see the `mixer` client making duplicate requests..._

**Jane**: _Check both logs for request ID
6E69546E-3CD8-4BED-9CE7-45CD3BF4B889. `mixer` sends that once, but `baker`
shows it arriving four times in quick succession. Only the fourth one
succeeds. That has to be retries._

Another pause.

**Julian**: _I’m an idiot. There’s a RetryPolicy for the whole namespace –
sorry, too many policies in the dashboard and I missed it. Deleting that since
you don’t want retries._

**Jane**: _Are you sure that’s a good–_

Jane’s phone shrills while she’s typing, and she drops it. When she picks it
up again she sees a stack of alerts. Quickly flipping through them, she feels
the blood drain from her face: there’s one for every single service in the
`baker` namespace.

**Jane**: _PUT IT BACK!!_

**Julian**: _Just did. Be glad you couldn't hear all the alarms here._ 😕

**Jane**: _What the hell just happened??_

**Julian**: _At a guess, all the workloads in the `baker` namespace actually
fail a lot, but they seem OK because there are retries across the whole
namespace?_ 🤔

Jane’s jaw drops.

**Jane**: _You’re saying that ALL of our services are broken??!_

**Julian**: _That’s what it looks like. Guessing your `baker` rollout would
have failed without retries turned on._

There is a pause while Jane thinks through increasingly unpleasant possibilities.

**Jane**: _I don't even know where to start here. How long did that
RetryPolicy go in? Is it the only thing like it?_

**Julian**: _I didn’t look closely before deleting it, but I think it said a
few months ago. And there are lots of different kinds of policy and lots of
individual policies, hang on a minute..._

**Julian**: _Looks like about 47 for your chunk of the world, a couple hundred
system-wide._

**Jane**: 😱 _Can you tell me what they’re doing for each of our services? I
can’t even_ look _at these things._ 😕

**Julian**: _That's gonna take awhile. Our tooling to show us which policies
bind to a given workload doesn't go the other direction._

**Jane**: _...wait. You have to_ build tools _to figure out basic configuration??_

Pause.

**Julian**: _Policy attachment is more complex than we’d like, yeah._ 😐
_Look, how ‘bout roll back your `baker` change for now? We can get together in
the morning and start sorting this out._

Jane shakes her head and rolls back her edits to the `baker` Deployment, then
sits looking out over the lake as the deployment progresses.

**Jane**: _Done. Are things happier now?_

**Julian**: _Looks like, thanks. Reckon you can get back to your sailboard._ 🙂

Jane sighs.

**Jane**: _Wish I could. Wind’s died down, though, and the sun is almost gone.
May as well head home._

One more look out at the lake.

**Jane**: _Thanks for the help. Wish we’d found better answers._ 😢

## The Proposal

The fundamental problem with policy attachment is that it **breaks the core
premise of Kubernetes as a declarative system**, because it’s not declarative:
it sets the world up for a sort of spooky action at a distance, to borrow
Einstein’s phrase. Policy attachment is not the only place where we see this
in Kubernetes, of course! but we submit that we shouldn't be adding any more
such places.

Given that the fundamental problem is that policy attachement isn't
declarative as written and should be made declarative, there is only one
fundamental answer: we need to modify the Kubernetes core resources to include
extension points where a given object refers to its modifier, rather than
having the modifying resource try to attach to its source. This is an ugly
job, but it’s the only way to deal with this situation.

This GEP proposes to start this process with the Gateway API resources.

## API

TODO: future iteration

## Questions and Answers

**Q**: _Why are you implying that there’s a problem with policy attachment?
Isn’t your parable really just showing us that Jane and Julian work for a
dysfunctional organization?_

**A**: As written, Evil Genius Cupcakes is _far_ from the most dysfunctional
organization I’ve seen. Jane and Julian support each other, neither casts
blame, both are clearly trying to do their best by the organization and their
customers even to their own cost. So the organization isn't really the
problem.

**Q**: _No organization would actually install a namespace-wide retry policy
and then forget about it, though._

**A**: I literally cannot even begin to count the number of times I’ve seen
something like this happen.

The most common scenario goes like this: it’s 8PM on a Friday and something
goes wrong. There is much screaming, wailing, and gnashing of teeth as the
on-call staff try to figure out what’s up. Inevitably, the SME is on vacation.
Someone suggests retries and they hastily slap in the CRD to enable them. The
post-mortem gets rescheduled a few times, and/or the person writing up the
timeline mistakenly notes that the retries were enabled for a given workload
rather than for the entire namespace, and no one ever figures out that error.
The post-mortem results in an action item of “fix this workload to not need
retries so we can turn retries off”, that goes into the backlog, and it gets
pushed down by more critical items.

That is a process problem for sure! but it's a sadly realistic one.

**Q**: _Okay, but in the real world, removing the RetryPolicy wouldn’t affect
every workload._

**A**: As soon as the namespace-wide RetryPolicy goes in, Jane’s team largely
loses the backstop of progressive rollout. As long as their workloads succeed
sometimes, progressive rollout has a good chance to succeed. After the few
months posited above, it’s not at all unlikely that every service will
actually be failing pretty often.

**Q**: _Fine. But in the real world, Jane would be able to see all the policy
objects herself, and this would be a non-issue._

**A**: Assuming permission to see everything necessary, please write me a
`kubectl` query to fetch every policy CRD that’s attached to an arbitrary
object. Remember to get policy CRDs attached to the enclosing namespace, too.

Challenging, no?

There’s a big difference between “having permission to see” and “being able to
effectively query and understand”. As policy attachment currently stands, you
need to be able to query many different kinds of CRDs _and_ filter them in a
couple of different ways that existing tooling isn't very good at.

**Q**: _Well then, in the real world, Jane would have access to higher-level
tools that know how to do that._

**A**: Those tools have yet to be written. Once they are, Jane and her team
will need to be taught that the tools exist and how to use them. From Jane’s
point of view, it's simpler not to need those tools: she'd rather just put the
right thing in her HTTPRoutes, and then be able to see them all when she reads
her HTTPRoutes.

**Q**: _What if we give Julian those tools? He could cope with them._

**A**: Sure, but now you’re back to a world in which Jane isn’t
self-sufficient and has to bottleneck on Julian. Neither of them will like
that.

**Q**: _Doesn't direct policy attachment make things better?_

**A**: Not really, no. Direct policy attachment is still spooky action at a
distance, so it doesn't really make things markedly better.

(That said, direct policy attachment _does_ sidestep a specific very
unpleasant scenario that I considered but didn’t write about. In that one,
Julian tries to tweak the RetryPolicy to disable the retries for just the
`baker` workload, but runs afoul of an override installed by Jasmine from the
cluster-ops team, which Julian doesn’t have permission to even see... so he
has to infer the existence of the override he can't see, and he can't do
anything about it.)

**Q**: _OK, so isn’t this really just a retry thing? It’s not like all
policies can affect things so broadly._

**A**: To state the obvious, the whole point of policy attachment is to set
policy -- and by definition, policy has very broad capabilities. Retry is
actually a fairly _narrow_ function: suppose the attached policy was instead a
WAF which was intentionally applied on every namespace (gotta protect
everything!), and Jasmine mistakenly changed its configuration? That could
affect everything in the entire cluster – possibly only a week after Jasmine
made the change, when the WAF gets an update that interacts poorly with the
configuration change.

**Q**: _Dude, c’mon. That’s Jasmine and the WAF shooting themselves in the
foot, not a problem with policy attachment._

**A**: You’re right that policy attachment didn’t cause the retry issue we
looked at first, nor would it cause the WAF problem above. What we're
concerned about is that policy attachement _does_ make it much harder for Jane
to understand what's happening so that she can fix it. That will have a real
impact on real people.

**Q**: _So you're just saying that everything is impossible and you're not
listening to my questions._

**A**: Well, most of your "questions" aren't questions! 🙂

And we definitely think it's possible to do something about the situation;
that's what this proposal is all about.
