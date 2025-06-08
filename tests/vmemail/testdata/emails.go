package testdata

const Email3 = `Delivered-To: loredana.cirstea@gmail.com
Received: by 2002:a5d:4522:0:b0:3a4:eadb:aeae with SMTP id j2csp1118498wra;
        Sat, 7 Jun 2025 14:02:46 -0700 (PDT)
X-Forwarded-Encrypted: i=4; AJvYcCXCeSOdbr0+cJHVGnPP3PvUrV+cpGL1rF8PtzEvjYbhNAxnLFAzY70KDwGx3tg0rQEKVYCfhXxGkFdn/E0aIVs0@gmail.com
X-Received: by 2002:a05:6102:1591:b0:4e6:245b:cf57 with SMTP id ada2fe7eead31-4e772a7ae98mr7820145137.24.1749330165756;
        Sat, 07 Jun 2025 14:02:45 -0700 (PDT)
ARC-Seal: i=3; a=rsa-sha256; t=1749330165; cv=pass;
        d=google.com; s=arc-20240605;
        b=HlpWCiVCmo/Ob+dNp1sGReo66eb6aQ7JA9PsGbgfM/k7BlZ8omusL7pUtFhP/FxAoM
         dSdLGmVxzbVkrE5Bqjs6BD0iq7HQNR+7Rdeg1LwhkMWvsVOkymRTlcu0vBohkJS10nbX
         Bi0H8aDvWnEBdp9aPtiOEEviBNanK+O2PFTvsxbOPWRs615n+BDIEEVgrLMG3Ds2bquE
         G+8sYKYZ5Rw5710QCejCubmRzepSLwS9/AEexEm5K1E3iXC7F1NojOTEpQaeJQnwWngg
         cPsD0L3B4YIXdkgg/tKu7eFHE7SHF5E9ZcM7DURaLGPqbjbgh3wtxf8I5A+CWB5Da68H
         auCw==
ARC-Message-Signature: i=3; a=rsa-sha256; c=relaxed/relaxed; d=google.com; s=arc-20240605;
        h=to:subject:message-id:date:from:mime-version:dkim-signature
         :delivered-to;
        bh=ojSfNd3Xfzhm7R18qUB+4Wi7dUR8WRE2XBS4Tw0JFvE=;
        fh=Y2OxNQeNb0EyUvov0Ehj0vqVnFaOcZdID/dvAs7c9SM=;
        b=CN4ljLJtkiDGX3d0N/gG26JNu831Zk6BxjaxcjL0iqufIpBO3SzQMvubPm9ufNcPV6
         Uv29+Z2tnpTM8Ly9dHtooLNkq1B0zsvFmM5x1Ma1MmJerhGywSFVS3rGC5j5L5vQSU2e
         w4J/nYmXUSV9G5ZcO6dIcCvNWS+05sf4+QA+WmWHinZaEPZrSMuI0Of170ykkgWydgdd
         gBFG0pqleBB0dpSqPAXBnxiLtR5fnjld+29/BSqm4NR3G72Gi5pDUp8MltonUte/82Hr
         m2+21dJc8bOUX1/fKz11XKysxyK1GWcODCgXf2UT0dMgga6COnzfsdEjHyd25R+ODd8z
         7XSA==;
        dara=google.com
ARC-Authentication-Results: i=3; mx.google.com;
       dkim=pass header.i=@gmail.com header.s=20230601 header.b=kfZ1QuWM;
       arc=pass (i=2 spf=pass spfdomain=gmail.com dkim=pass dkdomain=gmail.com dmarc=pass fromdomain=gmail.com);
       spf=pass (google.com: domain of quodum.one+caf_=loredana.cirstea=gmail.com@gmail.com designates 209.85.220.41 as permitted sender) smtp.mailfrom="quodum.one+caf_=loredana.cirstea=gmail.com@gmail.com";
       dmarc=pass (p=NONE sp=QUARANTINE dis=NONE) header.from=gmail.com;
       dara=pass header.i=@gmail.com
Return-Path: <quodum.one+caf_=loredana.cirstea=gmail.com@gmail.com>
Received: from mail-sor-f41.google.com (mail-sor-f41.google.com. [209.85.220.41])
        by mx.google.com with SMTPS id a1e0cc1a2514c-87ebd1f046fsor2017376241.4.2025.06.07.14.02.45
        for <loredana.cirstea@gmail.com>
        (Google Transport Security);
        Sat, 07 Jun 2025 14:02:45 -0700 (PDT)
Received-SPF: pass (google.com: domain of quodum.one+caf_=loredana.cirstea=gmail.com@gmail.com designates 209.85.220.41 as permitted sender) client-ip=209.85.220.41;
Authentication-Results: mx.google.com;
       dkim=pass header.i=@gmail.com header.s=20230601 header.b=kfZ1QuWM;
       arc=pass (i=2 spf=pass spfdomain=gmail.com dkim=pass dkdomain=gmail.com dmarc=pass fromdomain=gmail.com);
       spf=pass (google.com: domain of quodum.one+caf_=loredana.cirstea=gmail.com@gmail.com designates 209.85.220.41 as permitted sender) smtp.mailfrom="quodum.one+caf_=loredana.cirstea=gmail.com@gmail.com";
       dmarc=pass (p=NONE sp=QUARANTINE dis=NONE) header.from=gmail.com;
       dara=pass header.i=@gmail.com
ARC-Seal: i=2; a=rsa-sha256; t=1749330165; cv=pass;
        d=google.com; s=arc-20240605;
        b=lWbnxAX5Vl9Pc13R0Y8novO+LtEjN5IJ5vOIU89Ato4sMqAMDswjzfIUQE31ldT0SI
         bNPQ4ZU1T3/4XOTWJ5RCHDnDQNN846BCjtov2gAd+uEMrkhijI2FkpcWWlzQqq+eLKft
         EL4XsN7pB2D9oRt+xQoB4DovcFPsgLERxZOpEKKvExVrtu7oKDbkIw7v1LJmif44ByzP
         Or7TLDOpwFluOoGDn6Je5s7EHlPrNU/nd0eNDjNz0P2rXPsIjodpoxAObtrlGxung/zm
         nRFRDUP2tmDgqfUhqgpM1lAV36UWfKR2uJEcBZr9GTTUNYiQpTDyVkUhrA8EhPk+jpas
         Mb1g==
ARC-Message-Signature: i=2; a=rsa-sha256; c=relaxed/relaxed; d=google.com; s=arc-20240605;
        h=to:subject:message-id:date:from:mime-version:dkim-signature
         :delivered-to;
        bh=ojSfNd3Xfzhm7R18qUB+4Wi7dUR8WRE2XBS4Tw0JFvE=;
        fh=Y2OxNQeNb0EyUvov0Ehj0vqVnFaOcZdID/dvAs7c9SM=;
        b=QaFFJyjipOg7waRoEHkxiXFvR6K4m7Eg3Gxv6MiAUxcsc4x37dnYGMZDQQTtM3GLHr
         mtZA9KdeTrTOCOr8nSUMtC/JyqWNIb5zlOomuOByCAOZefqFwqip27fTrE1Uy7sM+p/A
         2l7nX5ZewQMSdil039Hr9Bfv6TMsOYEcQ/haKQ5yNNoc1QizxiIoJEXG0EwO/ALV10Ba
         9i2Y7+soTVa/oP1oUppM6bDfDNFr+2sHVS1jhgfehj93CYOU15CmAPv41XrcZxiAWwF1
         3YDQjC4YK9VdnRZgm7p7J+jkSP7R8d73+XUTmIREkcP9rsb/YRp5rPBheskJjOKzYRev
         rvDg==;
        dara=google.com
ARC-Authentication-Results: i=2; mx.google.com;
       dkim=pass header.i=@gmail.com header.s=20230601 header.b=kfZ1QuWM;
       spf=pass (google.com: domain of seth.one.info@gmail.com designates 209.85.220.41 as permitted sender) smtp.mailfrom=seth.one.info@gmail.com;
       dmarc=pass (p=NONE sp=QUARANTINE dis=NONE) header.from=gmail.com;
       dara=pass header.i=@gmail.com
X-Google-DKIM-Signature: v=1; a=rsa-sha256; c=relaxed/relaxed;
        d=1e100.net; s=20230601; t=1749330165; x=1749934965;
        h=to:subject:message-id:date:from:mime-version:dkim-signature
         :delivered-to:x-forwarded-for:x-forwarded-to:x-gm-message-state:from
         :to:cc:subject:date:message-id:reply-to;
        bh=ojSfNd3Xfzhm7R18qUB+4Wi7dUR8WRE2XBS4Tw0JFvE=;
        b=Y6OHssp72EjmqIPJtPgxZ//jPS5t7BcAfsNZiR1sV1ziIxHNx5ZEvoVUHJqBCCbO55
         8YSyidcTD3o6LZtb4U3WRyjtOxwNsXRA15URRisBPfpPcDIcg+qGct3NDNJY1lAfgSz9
         EaIwN5g6vfwSBaE3MDn6wLUeSWB/i7KXhYZ8JtSsKz+QM5QmT0XPX3jKWeWRmXuZrDsQ
         3jt4EZvsCaK1hPZt/nfrs9bei87Eh0jGq3XsDK4XJ6OmrboxT51W1/A7hV0njH1+NXne
         UCQqVlwrvAljY+J1RZrKTnfRHodWsY5iXD8VG0Q3c+a+IW2rKSMTkarAZ1BXdlWOy1pB
         qUGA==
X-Forwarded-Encrypted: i=2; AJvYcCV+MGVr8AbPUNPtkCENdKQ9/om4F7bh6rVCZ/U85mQdWQ1cO/DhtH3A3ZlKl0JmVbDwD6MYEGntDD5fu0qNc0aU@gmail.com
X-Gm-Message-State: AOJu0Yy8YaDDySl5i9Kx7Bb3ychbqvvvtLf/xgfORHoOyYrvw9P7JUvZ 6eeSGOXRexgzuhG8yAVj8KpFcQDR3L6te6Xb4mN9cE4WFkw2ps0Djs4Bolr7MbuTddyjOGI6iDy vOZvD1V33q5KCXP50AijwSxP/ZAc1o0mtQhbOqBIRHPYe9a9iVc9X1aRaoBl61AF1
X-Received: by 2002:a05:6102:50a2:b0:4e4:5df7:a10a with SMTP id ada2fe7eead31-4e772a2b47fmr6935625137.16.1749330165299;
        Sat, 07 Jun 2025 14:02:45 -0700 (PDT)
X-Forwarded-To: loredana.cirstea@gmail.com
X-Forwarded-For: quodum.one@gmail.com loredana.cirstea@gmail.com
Delivered-To: quodum.one@gmail.com
Received: by 2002:a05:612c:41e4:b0:4d5:d834:4912 with SMTP id lf36csp841528vqb;
        Sat, 7 Jun 2025 14:02:44 -0700 (PDT)
X-Received: by 2002:a17:903:188:b0:234:ef42:5d65 with SMTP id d9443c01a7336-23601dec50bmr102523575ad.52.1749330164188;
        Sat, 07 Jun 2025 14:02:44 -0700 (PDT)
ARC-Seal: i=1; a=rsa-sha256; t=1749330164; cv=none;
        d=google.com; s=arc-20240605;
        b=LhbQZpgTha5rFUHgT3Hx7x6Azp8YDMfivcGpr1KrmWM5HYhh5itoxpiNCbfJt/qqg/
         4i6kAbNcSwJTpteEqKKqfd7qGa3TsUABrmnLoISBsEBjBJrXKPAXk8xAxz3t7+4nTtEm
         akbPgOYEeZ1yf/iLFx2pVyjX1G5mkRPnr8Ki3vvnwGH9FeJqp1cpYohFC6ZJ+BwXgQee
         Kum7WoZpJe9dE1LB7IvbOZjf6lYqczYuTXYqvAp/5CiJ71VtWC7WCjaj70hgya2Ni7wL
         pd6O/7tOS7N1zdOZ3CDEQE8QKAUYjINydRPRl7FsBDR/70bz8J3d9UDZ0TLy8WWwHMl9
         PGxA==
ARC-Message-Signature: i=1; a=rsa-sha256; c=relaxed/relaxed; d=google.com; s=arc-20240605;
        h=to:subject:message-id:date:from:mime-version:dkim-signature;
        bh=ojSfNd3Xfzhm7R18qUB+4Wi7dUR8WRE2XBS4Tw0JFvE=;
        fh=n+IaoLt+5adQ8VuKzY0aXQQSV3ahwNDyeYNQr2RqzCA=;
        b=Enfk/PmJSw2nr4bDaXkvJGjOUFZqt0nSgyGNN91JvMHEJo8uA5ccNINrVURsLDljUL
         upqOnQIy6t+Fh8QdSrugp5QZnwga5XM+Irb5SDBLgEEGhwrMk+AGqxHotxd0hX9ECLvn
         jcpHqa+GE3VfjtTXwp6EVkdoB2kMHHIJJhxw1SGc58f6oyTL5vM4391kxarJCYiYKr7V
         4bVLdheQBvvr2FF6+ODkWsmY2HTY/mdWmZw6tyBfVklsN9RoP+sMKGsq6VL3STJkoiMI
         MMhG2vVqx2kLhorqcHXur6xrESUglwaC+QSCxj1XLR+RtFXk37NEZxpaRJnIleGzSLfl
         wqpw==;
        dara=google.com
ARC-Authentication-Results: i=1; mx.google.com;
       dkim=pass header.i=@gmail.com header.s=20230601 header.b=kfZ1QuWM;
       spf=pass (google.com: domain of seth.one.info@gmail.com designates 209.85.220.41 as permitted sender) smtp.mailfrom=seth.one.info@gmail.com;
       dmarc=pass (p=NONE sp=QUARANTINE dis=NONE) header.from=gmail.com;
       dara=pass header.i=@gmail.com
Return-Path: <seth.one.info@gmail.com>
Received: from mail-sor-f41.google.com (mail-sor-f41.google.com. [209.85.220.41])
        by mx.google.com with SMTPS id 98e67ed59e1d1-3134b1342a0sor2461302a91.8.2025.06.07.14.02.44
        for <quodum.one@gmail.com>
        (Google Transport Security);
        Sat, 07 Jun 2025 14:02:44 -0700 (PDT)
Received-SPF: pass (google.com: domain of seth.one.info@gmail.com designates 209.85.220.41 as permitted sender) client-ip=209.85.220.41;
DKIM-Signature: v=1; a=rsa-sha256; c=relaxed/relaxed;
        d=gmail.com; s=20230601; t=1749330164; x=1749934964; dara=google.com;
        h=to:subject:message-id:date:from:mime-version:from:to:cc:subject
         :date:message-id:reply-to;
        bh=ojSfNd3Xfzhm7R18qUB+4Wi7dUR8WRE2XBS4Tw0JFvE=;
        b=kfZ1QuWM+eRZHhMmCDCWY2OnAkPX9mXY45UIV9Cw6HhrRIgvdeA+ebkPbw8NkCae2r
         IO+5ZH3GYLrAU3qRk9R/KTowTcfYIEcedeS4jTKMG34ud8TjEshAtoo5rqWnkxeWp87x
         nI18igSZ2TTPxuQbq68hYwUr1umh/rVAnYmbwgyyfDgFSpDC/UNQNlMnzjfgS4oH9CdA
         TI4IyxKKckTjq+6Vd3EcnOp3ajrweqXU8Ddnj8vCZfC6RdK5jXR6W5GjtNFW36jZEizl
         9+ix1msFnaMlY4WEqwut9J/OxUzuSbdyObr+K65lJKwZbsf6OSouttgG2olEu0xPZcAN
         cwPg==
X-Gm-Gg: ASbGncv70TIozO1qSAqPL9krziF1HZRgVh9jmSMSr+NbUEgzTOXvU/DySuRirBkmdQA zGVCz1zyfXaVSp7rKL9/+En9j9F6faGGsIlhy8cRvGLC0Qcfl0AdAoiQvskS+8f2kQ+seiTS2e5 W76x7JlvqYOv7vSM3QqVJQSPwxBkU8P2E=
X-Google-Smtp-Source: AGHT+IGuCWJ622kjmZrgLcUdGqzZGKo9aJyL1tTWFZtX/gvVglNK2LpPNCdV1G6lcKUXfqUEMkL1TCxvf3nxGW0e6HA=
X-Received: by 2002:a17:90a:e7c1:b0:311:b3e7:fb31 with SMTP id 98e67ed59e1d1-31346c4c5admr12766932a91.0.1749330163695; Sat, 07 Jun 2025 14:02:43 -0700 (PDT)
MIME-Version: 1.0
From: Seth One <seth.one.info@gmail.com>
Date: Sat, 7 Jun 2025 23:02:32 +0200
X-Gm-Features: AX0GCFs36cT3L3H4caa3V7pF3rCbFNeR5ntF4SdO_W2X6mV-R74n-W5a1bbbn44
Message-ID: <CADMWPsWFe5PULwg5ymTHc0KZ4Ch-xbY-yAUaO-jzwcf1+yGp+w@mail.gmail.com>
Subject: Testing ARC
To: quodum.one@gmail.com
Content-Type: multipart/alternative; boundary="000000000000008442063701abbf"

--000000000000008442063701abbf
Content-Type: text/plain; charset="UTF-8"

Testing ARC

--000000000000008442063701abbf
Content-Type: text/html; charset="UTF-8"

<div dir="ltr">Testing ARC</div>

--000000000000008442063701abbf--`

const Email2 = `X-Mox-Reason: msgfromfull
Delivered-To: test@mail.provable.dev
Return-Path: <seth.one.info@gmail.com>
Authentication-Results: mail.provable.dev;
	iprev=pass (without dnssec) policy.iprev=209.85.215.175;
	dkim=pass (2048 bit rsa, without dnssec) header.d=gmail.com header.s=20230601
	header.a=rsa-sha256 header.b=UhsFIF+9QLwX;
	spf=pass (without dnssec) smtp.mailfrom=gmail.com;
	dmarc=pass (without dnssec) header.from=gmail.com
Received-SPF: pass (domain gmail.com) client-ip=209.85.215.175;
	envelope-from="seth.one.info@gmail.com"; helo=mail-pg1-f175.google.com;
	mechanism="include:_netblocks.google.com"; receiver=mail.provable.dev;
	identity=mailfrom
Received: from mail-pg1-f175.google.com ([209.85.215.175]) by
	mail.provable.dev ([85.215.130.119]) via tcp with ESMTPS id
	zls7D27Eslau1rXXZNMCKw (TLS1.3 TLS_AES_128_GCM_SHA256) for
	<test@mail.provable.dev>; 6 Jun 2025 16:12:10 +0000
Received: by mail-pg1-f175.google.com with SMTP id 41be03b00d2f7-7fd581c2bf4so1502455a12.3
        for <test@mail.provable.dev>; Fri, 06 Jun 2025 09:12:10 -0700 (PDT)
DKIM-Signature: v=1; a=rsa-sha256; c=relaxed/relaxed;
        d=gmail.com; s=20230601; t=1749226329; x=1749831129; darn=mail.provable.dev;
        h=to:subject:message-id:date:from:mime-version:from:to:cc:subject
         :date:message-id:reply-to;
        bh=vvLu1dIImn/XOQtsPkQDUQcze0x0yESF1ukjbKQaeVI=;
        b=UhsFIF+9QLwXINC7sZqgJKhWLuIa2tKlTn75qPQcE7LVZSoI/J/TIymYdBC94PFOng
         XDbu2B212v4ysxC1X7bracqdbE8sF7UxMOhHeWT6wW0RGqngZj9KoOsb9JQk4bwoDQ4N
         QtdDF6Gkv9V3Vnq+Fr3RkbtTCtUx1Hl4KCALvCRPHhW2+ERdMPkC+5x0an/oda0fPpm7
         nAyi58rBfil6XQfTJHmjcOByHxMltNVVZti3t4c21tI7BUC0A3gAo/CFuXo4H7x/NVC/
         3TO9S6smIKVVzJzLEOoN6JlGOeCmQ7IGgQDKEfzb1mb0sKM9bK2D6r3HEauppyV341sx
         f5Pg==
X-Google-DKIM-Signature: v=1; a=rsa-sha256; c=relaxed/relaxed;
        d=1e100.net; s=20230601; t=1749226329; x=1749831129;
        h=to:subject:message-id:date:from:mime-version:x-gm-message-state
         :from:to:cc:subject:date:message-id:reply-to;
        bh=vvLu1dIImn/XOQtsPkQDUQcze0x0yESF1ukjbKQaeVI=;
        b=gYd15MVJgwdlQkk228Gf6VqZc7Rb40wPal5n5nh+ktv5vR4l5ZzHmXO8JzEV2h8h8J
         xD0oqWcV8T6NV1UUx6d44J4JFxLXaL+eeQt7Conu9gxnpSo45ES3/NJlqX4lmlMkVQbQ
         Q1Fv0iEuDWYQgUW6pfUl3t3YsgUIqd92TD3mlsixohStSpgR7eA+8i5FIsM1kM7jaUmQ
         Fi9B0Lcni4I+2zWwn/ga1BOzQKPjYFDgw1iR/3LM1gNGxqrlrBZVD6WQkUQ5FkF7zvb9
         QtQJaYl1oPofUTqur1UyDaNrjoUTwar0f+TCOdPIEFXL8DQTiGlWtUq80/eVtR+uKanH
         toIQ==
X-Gm-Message-State: AOJu0Yxnx7lCi4lSKxzoFjucLsyc4ttUe6VsTdtq/SNacqQUeqPZ1IDk
	QVve9e6xPL7/NB99s0oRaSts+P2wh5f4H3hp1K+aab9SDlvOGH0Ajlj5a1hR3qy1X59XUqCdPup
	BHoeiG58d6B+aSvWLDbYDNKS15NMQjfZDMmyQt1Q=
X-Gm-Gg: ASbGncvqeuHjtwEolQgjp7NA2jrRLMuX+WsSA6M/LX2xR1ulMm52q4KdigXACgjPpPN
	hHfRH2YeOtFu0yOYE/wAVtSUH0Dh0bnO84NuUSQi/87o6n/JELQi9qIIZtOCsxsMFkd+jNVKain
	zYEWSChMiZgYAWtsq8jMwatcROyKOf8HE=
X-Google-Smtp-Source: AGHT+IEkBzg244c2Jv+tVOficNKeslzC1aq4I5tTtGtdijOrrwrraQ8y+UnoLOm2PJMpFu3jPMXx7cXnqz/o/1+vRZQ=
X-Received: by 2002:a17:90b:4a45:b0:311:a623:676c with SMTP id
 98e67ed59e1d1-3134767ec60mr6070787a91.27.1749226328813; Fri, 06 Jun 2025
 09:12:08 -0700 (PDT)
MIME-Version: 1.0
From: Seth One <seth.one.info@gmail.com>
Date: Fri, 6 Jun 2025 18:11:56 +0200
X-Gm-Features: AX0GCFs4YjxbyWZKShmsUYKcyQqJ1JDbsLjuZS_y7NyD-9jm55h-iQO7CSRehnw
Message-ID: <CADMWPsWixshaw8BGRN0THBZGwkxdfqTVu0E2_ZwNJ9pWCYjE0A@mail.gmail.com>
Subject: testing DKIM
To: test@mail.provable.dev
Content-Type: multipart/alternative; boundary="000000000000f5f7780636e97dcd"

--000000000000f5f7780636e97dcd
Content-Type: text/plain; charset="UTF-8"

testing DKIM

--000000000000f5f7780636e97dcd
Content-Type: text/html; charset="UTF-8"

<div dir="ltr">testing DKIM</div>

--000000000000f5f7780636e97dcd--`

const Email1 = `Delivered-To: seth.one.info@gmail.com
Received: by 2002:a05:7022:2201:b0:9b:65ec:421a with SMTP id bu1csp4328274dlb;
        Mon, 26 May 2025 12:40:48 -0700 (PDT)
X-Google-Smtp-Source: AGHT+IGR+KnDW62CVyj1GCKbaJECfByBnPYC9iWUN900XQRBBYDi1jQ5r+SfEuqUGxKYELny/oUi
X-Received: by 2002:a05:6000:2406:b0:3a4:e0e1:8dc7 with SMTP id ffacd0b85a97d-3a4e0e18f91mr219307f8f.55.1748288448523;
        Mon, 26 May 2025 12:40:48 -0700 (PDT)
ARC-Seal: i=1; a=rsa-sha256; t=1748288448; cv=none;
        d=google.com; s=arc-20240605;
        b=KtugwMnC2RDzjBybDCoWcne8FPhX80MnrGSbnhEOYN1PnFUQBFOOqQTAK9M+czTMaq
         ReI0SNMgimNT9O2aHqwo49qFIc8+1e+gcXTvV3kINM+pChrsPnST4P8iAs4qAfo34qqL
         ELt81cSas3nYLu64OXc7tTX8nj3kXBs5ciCBFN0rk1Oehyx520P4xQev3f/i217okXGI
         J9XtBkzrzUN+8MSkyjw6LvFjidoWUOFCjnc1IqtnakyBelq6TjsPAJY+8GWjwNqDqhGd
         RnXgYunjsKzCdm16PErCJ/B6rt6RFtKn59RgElhhksryrjQK7YZlb6wgHzbnHX13s7ye
         uVvQ==
ARC-Message-Signature: i=1; a=rsa-sha256; c=relaxed/relaxed; d=google.com; s=arc-20240605;
        h=content-transfer-encoding:mime-version:user-agent:date:message-id
         :subject:to:from:dkim-signature:dkim-signature;
        bh=7pO91KUgp6O1ny9F0HAz1889XmTtk0tlTKJf0BobX4I=;
        fh=B5nhD792HohsfLAnkzLmSID4s52GYCKkejwiXkLPDKM=;
        b=DAi++cbNRxnT8M3LKh4clc7Gvnu55KM+B8Pt+WOvjItZWzhap0dJW98c4sC6aoT+YN
         DbRqE0SCS1ah+7x6j2fbhx8gM7Omve9sC8ziUYF0pRIQ1uqzCaq3OYJb+UxaywpjJoM1
         yev+JD3F0TdzFlG5Dp0TwXZlzx++t1nQPtYqjXK5fKMQVUhtphzXzVbt1O3ujHIlkkeU
         xGUbzCaHfhQqZvOcXCpRPmrY9qB4vWLyYXbTW6vkn3EBYwqP1si3lPxLml89CElku1wW
         GfFkSxeDh7oFcbD1baYZCP7x0ZE3XRexxv4l+LV9V5/eWF3rt9F4F4vRI9CL92DDI4TT
         w8og==;
        dara=google.com
ARC-Authentication-Results: i=1; mx.google.com;
       dkim=neutral (no key) header.i=@mail.provable.dev header.s=2024a;
       dkim=pass header.i=@mail.provable.dev header.s=2024b header.b=QlLiPngW;
       spf=pass (google.com: domain of test@mail.provable.dev designates 85.215.130.119 as permitted sender) smtp.mailfrom=test@mail.provable.dev;
       dmarc=pass (p=REJECT sp=REJECT dis=NONE) header.from=mail.provable.dev
Return-Path: <test@mail.provable.dev>
Received: from mail.provable.dev (mail.provable.dev. [85.215.130.119])
        by mx.google.com with ESMTPS id ffacd0b85a97d-3a4cf025cfcsi3934300f8f.283.2025.05.26.12.40.48
        for <seth.one.info@gmail.com>
        (version=TLS1_3 cipher=TLS_AES_128_GCM_SHA256 bits=128/128);
        Mon, 26 May 2025 12:40:48 -0700 (PDT)
Received-SPF: pass (google.com: domain of test@mail.provable.dev designates 85.215.130.119 as permitted sender) client-ip=85.215.130.119;
Authentication-Results: mx.google.com;
       dkim=neutral (no key) header.i=@mail.provable.dev header.s=2024a;
       dkim=pass header.i=@mail.provable.dev header.s=2024b header.b=QlLiPngW;
       spf=pass (google.com: domain of test@mail.provable.dev designates 85.215.130.119 as permitted sender) smtp.mailfrom=test@mail.provable.dev;
       dmarc=pass (p=REJECT sp=REJECT dis=NONE) header.from=mail.provable.dev
Received: from mail.provable.dev by mail.provable.dev id 5sFAHppBc6epO9hJtpJ_ug for <seth.one.info@gmail.com>; 26 May 2025 19:40:47 +0000
DKIM-Signature: v=1; d=mail.provable.dev; s=2024a; i=test@mail.provable.dev; a=ed25519-sha256; t=1748288447; x=1748547647; h=From:To:Cc:Bcc:Reply-To: References:In-Reply-To:Subject:Date:Message-Id:Content-Type:From:To:Subject: Date:Message-Id:Content-Type; bh=7pO91KUgp6O1ny9F0HAz1889XmTtk0tlTKJf0BobX4I=; b=zvoQEotj0nrwSfZsa733b61T8p DUrrcSjzM2HJS1cjgyOBfoXhcSEQ3Pgz03Ro1eJyPYu9ZS3Lav9zwAeJCnBA==
DKIM-Signature: v=1; d=mail.provable.dev; s=2024b; i=test@mail.provable.dev; a=rsa-sha256; t=1748288447; x=1748547647; h=From:To:Cc:Bcc:Reply-To: References:In-Reply-To:Subject:Date:Message-Id:Content-Type:From:To:Subject: Date:Message-Id:Content-Type; bh=7pO91KUgp6O1ny9F0HAz1889XmTtk0tlTKJf0BobX4I=; b=QlLiPngWTmYFj8AmjffZjScSER eMMiqk2KNrafqx7nsnc3hISCidHbGk0fwB5RMFVb8u/criLgUnOfbIQy/6/WXaozG3Q6BFTrd98sE fM8B/ip0hLgGzXgWMqPDRWlVDyG1aE80jFapKUk/iJfdJIec//0uDaIp09QREh2Iyt4LE2zxjjd9T swvtjS37aIwAFvI+w1Ww0iKnWP+4wG64KFAsV7FcEWUdz43RIGAxqeMreEACBhFc5qoO6oO1aNDGq AYTejHo6pw9YL7TJJh/Dt49te2zwdZQIR5TDVKLIyifyBwg4NOUUoeAhuLsok6NgIRWnirZGcT34z 2pFaRoPA==
From: <test@mail.provable.dev>
To: Seth One <seth.one.info@gmail.com>
Subject: a new message
Message-Id: <X2Ct4oFU2a7Kvf9lrkCQvg@mail.provable.dev>
Date: 26 May 2025 19:40:47 +0000
User-Agent: moxwebmail/v0.0.10
MIME-Version: 1.0
Content-Type: text/plain; charset=us-ascii
Content-Transfer-Encoding: 7bit

a new message`
