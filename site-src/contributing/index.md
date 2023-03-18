# How to Get Involved

This page contains links to all of the meeting notes, design docs and related
discussions around the APIs. If you're interested in working towards a formal
role in the project, refer to the [Contributor
Ladder](/contributing/contributor-ladder).

## Feedback and Questions

For general feedback, questions or to share ideas please feel free to [create a
new discussion][gh-disc].

[gh-disc]:https://github.com/kubernetes-sigs/gateway-api/discussions/new

## Bug Reports

Bug reports should be filed as [Github Issues][gh-issues] on this repo.

**NOTE**: If you're reporting a bug that applies to a specific implementation of
Gateway API and not the API specification itself, please check our
[implementations page][implementations] to find links to the repositories where
you can get help with your specific implementation.

[gh-issues]: https://github.com/kubernetes-sigs/gateway-api/issues/new/choose
[implementations]:https://gateway-api.sigs.k8s.io/implementations/

## Communications

Major discussions and notifications will be sent on the [SIG-NETWORK mailing
list][signetg].

We also have a [Slack channel (sig-network-gateway-api)][slack] on k8s.io for day-to-day
questions, discussions.

[signetg]: https://groups.google.com/forum/#!forum/kubernetes-sig-network
[slack]: https://kubernetes.slack.com/archives/CR0H13KGA

## Meetings

Gateway API has multiple meetings covering different aspects and topics of the
Gateway API project. The following calendar includes _all_ SIG Network meetings
(which therefore includes all Gateway API meetings which include "Gateway API"
somewhere in their name):

<iframe
  src="https://calendar.google.com/calendar/embed?src=88fe1l3qfn2b6r11k8um5am76c%40group.calendar.google.com"
  style="border: 0" width="800" height="600" frameborder="0"
  scrolling="no">
</iframe>

### Main Meeting

The main Gateway API community meetings happen weekly on Mondays at 3pm Pacific
Time (23:00 UTC):

* [Zoom link](https://zoom.us/j/441530404)
* [Convert to your timezone](http://www.thetimezoneconverter.com/?t=15:00&tz=PT%20%28Pacific%20Time%29)
* [Add to your calendar](https://calendar.google.com/event?action=TEMPLATE&tmeid=NXU4OXYyY2pqNzEzYzUwYnVsYmZwdXJzZDlfMjAyMTA1MTBUMjIwMDAwWiA4OGZlMWwzcWZuMmI2cjExazh1bTVhbTc2Y0Bn&tmsrc=88fe1l3qfn2b6r11k8um5am76c%40group.calendar.google.com&scp=ALL)

Being the main meeting for Gateway API, the topics can vary here and often this
is where new topics and ideas are discussed. However if you're simply
interested in the ingress use case, this is the common forum for that.

### Mesh Meeting

[GAMMA](/contributing/gamma/) is the initative within Gateway API to use the
resources provided by Gateway API for service mesh use cases. Meetings happen
weekly on Tuesdays, alternating between 3pm Pacific Time (23:00 UTC) and 8AM
Pacific Time (16:00 UTC):

* [Zoom link](https://zoom.us/j/96951309977)
* Convert to your timezone: [3pm PT](http://www.thetimezoneconverter.com/?t=15:00&tz=PT%20%28Pacific%20Time%29)/[8am PT](http://www.thetimezoneconverter.com/?t=08:00&tz=PT%20%28Pacific%20Time%29)

### Code Jam Meeting

The Gateway API "Code Jam" is less of a meeting and more of a hangout to discuss
and pair on Gateway API and technologies relevant to the project. This is an
open agenda meeting (feel free to bring your topics) where the following kinds
of activities (focusing on Gateway API related things) are encouraged:

- Brainstorming
- Code pairing
- Demos
- Getting help

This meeting puts an emphasis on being fun and laid back: If you're looking to
build further consensus and progress a [GEP][geps] then the _main_ meeting is
likely the place you'll want to bring your topic. However, if you're working on
adding [conformance tests][conformance] to your downstream implementation and
having some trouble and want some help getting things working, this meeting is a
good place for that kind of topic.

* [Zoom link](https://zoom.us/j/96900767253)
* [8:30am PT](http://www.thetimezoneconverter.com/?t=08:30&tz=PT%20%28Pacific%20Time%29)

[geps]:https://gateway-api.sigs.k8s.io/geps/overview/

### Meeting Notes and Recordings

Meeting agendas and notes are maintained in the [meeting notes
doc][meeting-notes]. Likewise, GAMMA has its own [meeting notes
doc][gamma-meeting-notes]. Feel free to add topics for discussion at an upcoming
meeting.

All meetings are recorded and automatically uploaded to the [Gateway API meetings Youtube playlist][gateway-api-yt-playlist].

#### Early Meetings
Some early community meetings were uploaded to a [separate YouTube
playlist][early-yt-playlist], and then to the [SIG Network Youtube playlist][sig-net-yt-playlist].

#### Initial Design Discussions

* [Kubecon 2019 San Diego: API evolution design discussion][kubecon-2019-na-design-discussion]
* [SIG-NETWORK: Ingress Evolution Sync][sig-net-2019-11-sync]
* [Kubecon 2019 Barcelona: SIG-NETWORK discussion (general topics, includes V2)][kubecon-2019-eu-discussion]

[gateway-api-yt-playlist]: https://www.youtube.com/playlist?list=PL69nYSiGNLP1GgO7k02ipPGZUFpSzGaHH
[sig-net-yt-playlist]: https://www.youtube.com/playlist?list=PL69nYSiGNLP2E8vmnqo5MwPOY25sDWIxb
[early-yt-playlist]: https://www.youtube.com/playlist?list=PL7KjrPTDcs4Xe6SZj-51WvBfufKf-la1O
[kubecon-2019-na-design-discussion]: https://docs.google.com/document/d/1l_SsVPLMBZ7lm_T4u7ZDBceTTUY71-iEQUPWeOdTAxM/preview
[kubecon-2019-eu-discussion]: https://docs.google.com/document/d/1n8AaDiPXyZHTosm1dscWhzpbcZklP3vd11fA6L6ajlY/preview
[sig-net-2019-11-sync]: https://docs.google.com/document/d/1AqBaxNX0uS0fb_fSpVL9c8TmaSP7RYkWO8U_SdJH67k/preview
[meeting-notes]: https://docs.google.com/document/d/1eg-YjOHaQ7UD28htdNxBR3zufebozXKyI28cl2E11tU/edit
[gamma-meeting-notes]: https://docs.google.com/document/d/1s5hQU0CB9ehjFukRmRHQ41f1FA8GX5_1Rv6nHW6NWAA/edit#

## Presentations and Talks

| Date           | Title |    |
|----------------|-------|----|
| November, 2019 | [Kubecon 2019 San Diego: Evolving the Kubernetes Ingress APIs to GA and Beyond][2019-kubecon-na-slides] | [slides][2019-kubecon-na-slides], [video][2019-kubecon-na-video]|
| November, 2019 | Kubecon 2019 San Diego: SIG-NETWORK Service/Ingress Evolution Discussion | [slides][2019-kubecon-na-community-slides] |
| May, 2019      | [Kubecon 2019 Barcelona: Ingress V2 and Multicluster Services][2019-kubecon-eu] | [slides][2019-kubecon-eu-slides], [video][2019-kubecon-eu-video]|
| March, 2018    | SIG-NETWORK: Ingress user survey | [data][survey-data], [slides][survey-slides] |

[2019-kubecon-na]: https://kccncna19.sched.com/event/UaYG/evolving-the-kubernetes-ingress-apis-to-ga-and-beyond-christopher-m-luciano-ibm-bowei-du-google
[2019-kubecon-na-slides]: https://static.sched.com/hosted_files/kccncna19/a5/Kubecon%20San%20Diego%202019%20-%20Evolving%20the%20Kubernetes%20Ingress%20APIs%20to%20GA%20and%20Beyond%20%5BPUBLIC%5D.pdf
[2019-kubecon-na-video]: https://www.youtube.com/watch?v=cduG0FrjdJA
[2019-kubecon-eu]: https://kccnceu19.sched.com/event/MPb6/ingress-v2-and-multicluster-services-rohit-ramkumar-bowei-du-google
[2019-kubecon-eu-slides]: https://static.sched.com/hosted_files/kccnceu19/97/%5Bwith%20speaker%20notes%5D%20Kubecon%20EU%202019_%20Ingress%20V2%20%26%20Multi-Cluster%20Services.pdf
[2019-kubecon-eu-video]: https://www.youtube.com/watch?v=Ne9UJL6irXY&t=1s
[survey-data]: https://github.com/bowei/k8s-ingress-survey-2018
[survey-slides]: https://github.com/bowei/k8s-ingress-survey-2018/blob/master/survey.pdf
[2019-kubecon-na-community-slides]: https://docs.google.com/presentation/d/1s0scrQCCFLJMVjjGXGQHoV6_4OIZkaIGjwj4wpUUJ7M

## Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of
Conduct](https://github.com/kubernetes/community/blob/master/code-of-conduct.md)
