enum CURL_RESULT {
	UNKNOWN,
	OK
};

CURL_RESULT curlGet() {
	this->updateStatus("Checking for updates...", false);

	CURL *c = curl_easy_init();

	struct curl_slist *headers = curl_slist_append(NULL, "Accept: text/html");

	curl_easy_setopt(c, CURLOPT_URL, "http://tydus.net/example.txt");
	curl_easy_setopt(c, CURLOPT_HTTPHEADER, headers);
	curl_easy_setopt(c, CURLOPT_TIMEOUT, 4*1000);
	curl_easy_setopt(c, CURLOPT_NOBODY, 1);

	CURLcode res = curl_easy_perform(c);

	curl_slist_free_all(headers);

	if (res == CURLE_OK) {
		return OK;
	} else {
		int responseCode{};

		cout << curl_easy_getinfo(c, CURLINFO_RESPONSE_CODE, &responseCode);

		return UNKNOWN;
	}
}
