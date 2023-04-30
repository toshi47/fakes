
function doRequest(url, request, goodLocation) {
    fetch(url, request)
        .then(response => {
            response.text().then(text => {
                try {
                    const data = JSON.parse(text)
                    let msg = "this is "
                    if (data["is_fake"] == false) {
                        msg += "not "
                    }
                    msg += "fake";
                    if (data["probability"] != null) {
                        msg += " with probability " + data["probability"].toFixed(2)
                    }
                    document.getElementById("result").textContent = msg
                } catch (err) {
                    document.getElementById("result").textContent = text
                }
            })

            if (!response.ok) {
                if (response.status == 401) {
                    window.location.href = "login.html"
                }
                return
            }

            if (goodLocation != null) {
                window.location.href = goodLocation
            }
        })
        .catch(error => {
            console.error(error.message);
            document.getElementById("result").textContent = error.message;
        });
}

function auth() {
    doRequest("/auth", {
        method: "POST"
    }, null)
}

function submitLoginForm(event) {
    event.preventDefault();
    const loginData = {
        username: document.getElementById("username").value,
        password: document.getElementById("password").value
    };
    doRequest("/auth/login", {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify(loginData)
    }, "/confirm.html")
}

function submitConfirmationForm(event) {
    event.preventDefault();
    const confirmationData = {
        code: document.getElementById("code").value,
    };
    doRequest("/auth/confirm", {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify(confirmationData)
    }, "/")
}

function submitPredictText(event) {
    event.preventDefault();
    const predictData = {
        data: document.getElementById("text").value,
    };
    doRequest("/predict_text", {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify(predictData)
    }, null)
}

function submitPredictLink(event) {
    event.preventDefault();
    const predictData = {
        data: document.getElementById("link").value,
    };
    doRequest("/predict_link", {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify(predictData)
    }, null)
}

function previewImage(event) {
    if (!event.target.files || !event.target.files[0]) {
        return
    }
    const file = document.getElementById("image_file").files[0];
    const reader = new FileReader();
    reader.onload = function () {
        document.getElementById("preview").src = reader.result
        document.getElementById("preview").style.display = "block";
        document.getElementById("result").textContent = ""
    };
    reader.readAsDataURL(file)
}

function submitPredictImage(event) {
    event.preventDefault();
    const file = document.getElementById("image_file").files[0];
    const reader = new FileReader();
    reader.onload = function () {
        const predictData = {
            data: document.getElementById("preview").src,
        };
        doRequest("/predict_image", {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify(predictData)
        }, null, null)
    };
    reader.readAsBinaryString(file)
}

function validatePassword(event) {
    event.preventDefault();
    if (document.getElementById("password").value != document.getElementById("password_confirmation").value) {
        document.getElementById("result").textContent = "passwords don't match!";
        return false;
    }
    document.getElementById("result").textContent = "";
    return true;
}

function submitRegisterForm(event) {
    event.preventDefault();
    if (!validatePassword(event)) {
        return;
    }
    const registerData = {
        username: document.getElementById("username").value,
        password: document.getElementById("password").value,
        email: document.getElementById("email").value
    };
    doRequest("/auth/register", {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify(registerData)
    }, "login.html")
}