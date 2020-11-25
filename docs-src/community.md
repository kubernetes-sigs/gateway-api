# How to contribute

This page contains links to all of the meeting notes, design docs and related
discussions around the APIs.

## Feedback and Bug Reports

Feedback and bug reports should be filed as [Github Issues][gh-issues] on this repo.

[gh-issues]: https://github.com/kubernetes-sigs/service-apis/issues/new/choose

## Communications

Major discussions and notifications will be sent on the [SIG-NETWORK mailing
list][signetg].

We also have a [Slack channel (sig-network-service-apis)][slack] on k8s.io for day-to-day
questions, discussions.

[signetg]: https://groups.google.com/forum/#!forum/kubernetes-sig-network
[slack]: https://kubernetes.slack.com/archives/CR0H13KGA

## Meetings

Meetings discussing the evolution of the service APIs will alternate times to
accommodate participants from various time zones. This calendar includes all
Service APIs meetings as well as any other SIG-Network meetings.

<iframe
  src="https://calendar.google.com/calendar/embed?src=88fe1l3qfn2b6r11k8um5am76c%40group.calendar.google.com"
  style="border: 0" width="800" height="600" frameborder="0"
  scrolling="no">
</iframe>

* Wednesday 11am Pacific (7pm UTC) <a target="_blank" href="https://zoom.us/j/441530404">[Zoom Link]</a> <a target="_blank" href="https://calendar.google.com/event?action=TEMPLATE&tmeid=MjUxOHEzdXNpZWwxYmRmYXFidGltbWU1dWFfMjAyMDEyMDJUMTkwMDAwWiA4OGZlMWwzcWZuMmI2cjExazh1bTVhbTc2Y0Bn&tmsrc=88fe1l3qfn2b6r11k8um5am76c%40group.calendar.google.com&scp=ALL"><img border="0" src="https://www.google.com/calendar/images/ext/gc_button1_en.gif"></a>

### Meeting notes

Meeting agendas and notes are recorded in the [meeting notes doc][meeting-notes].

All meetings are recorded and automatically uploaded to youtube:
[Meeting recordings](https://www.youtube.com/playlist?list=PL69nYSiGNLP2E8vmnqo5MwPOY25sDWIxb).

Some initial recordings of this working group were done manually and can be
found in the below table:

| Date               |                                |
|--------------------|--------------------------------|
| Future meetings    | Check the calendar             |
| February 27, 2019  | [meeting notes][meeting-notes], [recording](https://youtu.be/QoInpFSTQbQ)   |
| February 20, 2019  | [meeting notes][meeting-notes], [recording](https://youtu.be/i_oDQuPEhd8)   |
| February 13, 2019  | [meeting notes][meeting-notes], [recording](https://youtu.be/bdFLubKi9_0)   |
| February 6, 2019   | [meeting notes][meeting-notes], [recording](https://youtu.be/XvpCFaTrtBA)   |
| January 30, 2019   | [meeting notes][meeting-notes], [recording](https://youtu.be/cTTqIR3muGk)   |
| January 23, 2019   | [meeting notes][meeting-notes], recording TODO |
| January 16, 2019   | [meeting notes][meeting-notes], [recording](https://youtu.be/ydA-epcZJQo)   |
| January 9, 2019    | [meeting notes][meeting-notes], [recording](https://youtu.be/C3zO67lXGrg)   |
| January 2, 2020    | [meeting notes][meeting-notes], recording didn't work :-( look at the notes |
| December 19, 2019  | [meeting notes][meeting-notes], [recording](https://youtu.be/FIcySpPkGa4)   |
| November, 2019     | [Kubecon 2019 San Diego: API evolution design discussion][kubecon-2019-na-design-discussion] |
| November, 2019     | [SIG-NETWORK: Ingress Evolution Sync][sig-net-2019-11-sync] |
| May, 2019          | [Kubecon 2019 Barcelona: SIG-NETWORK discussion (general topics, includes V2)][kubecon-2019-eu-discussion] |

[kubecon-2019-na-design-discussion]: https://docs.google.com/document/d/1l_SsVPLMBZ7lm_T4u7ZDBceTTUY71-iEQUPWeOdTAxM/preview
[kubecon-2019-eu-discussion]: https://docs.google.com/document/d/1n8AaDiPXyZHTosm1dscWhzpbcZklP3vd11fA6L6ajlY/preview
[sig-net-2019-11-sync]: https://docs.google.com/document/d/1AqBaxNX0uS0fb_fSpVL9c8TmaSP7RYkWO8U_SdJH67k/preview
[meeting-notes]: https://docs.google.com/document/d/1eg-YjOHaQ7UD28htdNxBR3zufebozXKyI28cl2E11tU/edit

## Presentations, Talks

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
