import scrapy


class PhpFileSpider(scrapy.Spider):
    name = "php_downloader"

    def start_requests(self):
        url = "https://example.com/download.php?id=123"
        yield scrapy.Request(url, callback=self.save_file)

    def save_file(self, response):
        filename = "downl"
        with open(filename, "wb") as f:
            f.write(response.body)
        self.log(f"Saved file as {filename}")
