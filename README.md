# PSEDB

![GitHub commit activity (branch)](https://img.shields.io/github/commit-activity/t/reinhart1010/psedb?authorFilter=1010bots&logo=github&label=Current%20DB%20Version%20(GitHub)&labelColor=%23004B77&color=%230095E1&link=https%3A%2F%2Fgithub.com%2Freinhart1010%2Fpsedb%2Fcommits%2Fmain)

[<img alt="PSE Badge" src="/pse-badge-en.png" height="60" width="180">](https://pse.kominfo.go.id/tdpse-detail/15516)

Quick Links:

+ [Official Website](https://psedb.reinhart1010.id)
+ [GitHub Repository](https://github.com/reinhart1010/psedb)

Welcome to PSEDB, an unofficial, developer-friendly dataset of Indonesia's registered [Private Sector Electronic System Operators (PSE/ESO)](https://pse.kominfo.go.id) and operating sites. This data may also be used for existing PSE/ESO-compliant web services for filtering or moderating external links and content. For example, when showing comments from external sites.

<img alt="Example dataset usage" src="/example-usage.png" height="413" width="421">

This API service is static (not dynamic) and must be updated regularly to reflect changes in the official PSE/ESO dataset provided by the Ministry of Telecommunication and Informatics, Republic of Indonesia.

If you have just set up your local copy of PSEDB through our Git repository, you can also fetch the latest version by running git pull regularly. However, those who prefer using Go may simply run go run main.go on the repository's root directory.

Our official version of the dataset, as hosted on https://psedb.reinhart1010.id and our Git repository, is currently automated using GitHub Actions, GitLab CI, and a homegrown CI service. You may also support us through [GitHub Sponsors](https://github.com/sponsors/reinhart1010), [Nih Buat Jajan](https://nihbuatjajan.com/reinhart1010), and [Saweria](https://saweria.co/reinhart1010).

## Using the API

To fetch the data, perform a GET request to `/url/{reversed domain name with slash instead of dot}.json`, such as `/url/id/reinhart1010/psedb.json` for `psedb.reinhart1010.id`.

Please also beware of sites who are using the `www` subdomain, since we treat them explicitly as different site than the root domain. For example, if `reinhart1010.id` on `/url/id/reinhart1010.json` returns a 404 - Not Found response, try checking `/url/id/reinhart1010/www.json` for `www.reinhart1010.id`.

You may also test out the REST API [directly from our website](https://psedb.reinhart1010.id#usage).

## Copyright and License

**Official ESO/PSE Dataset © Ministry of Communication and Informatics of Republic of Indonesia. All rights reserved.**

In accordance to Republic of Indonesia copyright laws (*Undang-Undang Nomor 28 Tahun 2014 Tentang Hak Cipta*), the usage of copyrighted content is exempted from infringement (Fair Use) for legal administrative and security purposes with proper attribution to the source and/or author(s).

**PSEDB © 2023 Reinhart Previano Koentjoro. Some rights reserved.**

Our official version of the dataset, as hosted on https://api.psedb.reinhart1010.id and our Git repositories, is licensed under [Open Database License (ODbL) 1.0](https://opendatacommons.org/licenses/odbl/1-0/). Our source code are released under [MIT License](/LICENSE).
