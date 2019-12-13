<!--
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
-->

# How to contribute

This page contains links to all of the meeting notes, design docs and related
discussions around the APIs.

## Meetings

Meetings discussing the evolution of the service APIs will alternate times to
accommodate participants from various time zones:

* Thursday 10:30 AM Pacific ([calendar link][cal-meeting-am])
* Thursday  4:30 (16:30) PM Pacific ([calendar link][cal-meeting-pm])

[cal-meeting-am]: https://calendar.google.com/event?action=TEMPLATE&tmeid=M21yb3U1YTcwbzJwdG0zN3IwMXFnZmg5ZDBfMjAyMDAxMTZUMTgzMDAwWiBzbGpwY3NsNzR2Zmhla292Y2NiMWZzdGxqY0Bn&tmsrc=sljpcsl74vfhekovccb1fstljc%40group.calendar.google.com&scp=ALL
[cal-meeting-pm]: https://calendar.google.com/event?action=TEMPLATE&tmeid=NmE1YXFtaHMzbzdsc3RyaDlzdDBta2NnZjdfMjAyMDAxMzFUMDAzMDAwWiBzbGpwY3NsNzR2Zmhla292Y2NiMWZzdGxqY0Bn&tmsrc=sljpcsl74vfhekovccb1fstljc%40group.calendar.google.com&scp=ALL

### Video conferencing

[meeting]: https://zoom.us/j/931530074?pwd=SmtZai9UeS9wVGE3azdFWGEwRTVtdz09

[Meeting link][meeting]

```
Topic: [SIG-NETWORK] Ingress/Service API evolution
Time: This is a recurring meeting Meet anytime

Join Zoom Meeting
https://zoom.us/j/931530074?pwd=SmtZai9UeS9wVGE3azdFWGEwRTVtdz09

Meeting ID: 931 530 074
Password: 621590

One tap mobile
+14086380968,,931530074# US (San Jose)
+16465588656,,931530074# US (New York)

Dial by your location
        +1 408 638 0968 US (San Jose)
        +1 646 558 8656 US (New York)
Meeting ID: 931 530 074
Find your local number: https://zoom.us/u/abdiTq5bx
```

* Join Zoom Meeting:
     * [https://zoom.us/j/931530074?pwd=SmtZai9UeS9wVGE3azdFWGEwRTVtdz09][meeting]
     * Meeting ID: 931 530 074
     * Password: 621590
     * One tap mobile
          * +14086380968,,931530074# US (San Jose)
          * +16465588656,,931530074# US (New York)
     * Dial by your location:
          * +1 408 638 0968 US (San Jose)
          * +1 646 558 8656 US (New York)
          * Find your local number: [https://zoom.us/u/abdiTq5bx](https://zoom.us/u/abdiTq5bx)

### Meeting notes

| Date |    |
|------|----|
| November, 2019 | [Kubecon 2019 San Diego: API evolution design discussion][kubecon-2019-na-design-discussion] |
| November, 2019 | [SIG-NETWORK: Ingress Evolution Sync][sig-net-2019-11-sync] |
| May, 2019      | [Kubecon 2019 Barcelona: SIG-NETWORK discussion (general topics, includes V2)][kubecon-2019-eu-discussion] |

[kubecon-2019-na-design-discussion]: https://docs.google.com/document/d/1l_SsVPLMBZ7lm_T4u7ZDBceTTUY71-iEQUPWeOdTAxM/preview
[kubecon-2019-eu-discussion]: https://docs.google.com/document/d/1n8AaDiPXyZHTosm1dscWhzpbcZklP3vd11fA6L6ajlY/preview
[sig-net-2019-11-sync]: https://docs.google.com/document/d/1AqBaxNX0uS0fb_fSpVL9c8TmaSP7RYkWO8U_SdJH67k/preview

## Design docs

| Title | Description |
|-------|-------------|
| [API sketch][api-sketch] | Sketch of the proposed API |

[api-sketch]:  https://docs.google.com/document/d/1BxYbDovMwnEqe8lj8JwHo8YxHAt3oC7ezhlFsG_tyag

## Presentations, Talks

| Date | Title |    |
|------|-------|----|
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
